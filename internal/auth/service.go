package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"votespher/pkg"

	"gorm.io/gorm"
)

// ผลลัพธ์หลังยืนยัน OTP สำเร็จ
type OTPConfirmResult struct {
	Token string
	Role  string
}

// ConfirmOTP ยืนยันรหัส OTP แล้วคืน JWT token
func ConfirmOTP(db *gorm.DB, req OTPConfirmRequest) (*OTPConfirmResult, error) {
	// 1. หา OTP จาก ref_code (เช็คว่าไม่หมดอายุและยังไม่ถูกใช้)
	otp, err := FindOTPByRefCode(db, req.RefCode)
	if err != nil {
		return nil, errors.New("ref_code ไม่ถูกต้องหรือ OTP หมดอายุแล้ว")
	}

	// 2. เช็คว่า otp_code ตรงกับที่บันทึกไว้ไหม
	if otp.OTPCode != req.OTPCode {
		// นับจำนวนครั้งที่กรอกผิด — block หลัง 5 ครั้ง
		newAttempts := otp.Attempts + 1
		update := map[string]interface{}{"attempts": newAttempts}
		if newAttempts >= 5 {
			update["is_used"] = true
			db.Model(otp).Updates(update)
			return nil, errors.New("กรอก OTP ผิดเกิน 5 ครั้ง รหัสถูกยกเลิกแล้ว กรุณาขอ OTP ใหม่")
		}
		db.Model(otp).Updates(update)
		return nil, fmt.Errorf("รหัส OTP ไม่ถูกต้อง (ครั้งที่ %d/5)", newAttempts)
	}

	// 3. mark OTP ว่าใช้แล้ว (กันไม่ให้ใช้ซ้ำ)
	if err := MarkOTPAsUsed(db, otp.ID); err != nil {
		return nil, err
	}

	// 4. ดึงข้อมูล Voter เพื่อเอา area_id
	voter, err := FindVoterByID(db, otp.VoterID)
	if err != nil {
		return nil, err
	}

	// 5. ตรวจว่าเป็น admin หรือเปล่า → กำหนด role อัตโนมัติ
	role := "voter"
	if IsAdmin(db, voter.ID) {
		role = "admin"
	}

	// 6. สร้าง JWT Token
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		return nil, errors.New("ระบบผิดพลาด กรุณาติดต่อผู้ดูแล")
	}
	token, err := pkg.GenerateToken(voter.ID, voter.AreaID, role, secretKey)
	if err != nil {
		return nil, err
	}

	return &OTPConfirmResult{Token: token, Role: role}, nil
}

// generateCitizenIDHash เอาไว้ hash citizen_id ก่อนบันทึกลง DB
// ใช้ใน /voter/verify endpoint
func generateCitizenIDHash(citizenID string) string {
	secretKey := []byte(os.Getenv("HASH_SECRET_KEY"))
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(citizenID))
	return hex.EncodeToString(h.Sum(nil))
}

// สุ่มเลข 6 หลัก สำหรับ OTP
func generateRandomOTP() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// สุ่มตัวอักษร 6 ตัว สำหรับ Ref Code
func generateRefCode() (string, error) {
	b := make([]byte, 3)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
