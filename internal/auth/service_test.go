package auth

import (
	"os"
	"testing"
)

// 1. ทดสอบการสุ่ม OTP (ต้องได้ 6 หลักเสมอ)
func TestGenerateRandomOTP(t *testing.T) {
	otp, err := generateRandomOTP()
	if err != nil {
		t.Fatalf("สุ่ม OTP พัง: %v", err)
	}

	if len(otp) != 6 {
		t.Errorf("OTP ต้องมี 6 หลัก แต่ได้มา: %s", otp)
	}
}

// 2. ทดสอบการสุ่ม Ref Code (ต้องได้ 6 ตัวอักษร hex)
func TestGenerateRefCode(t *testing.T) {
	ref, err := generateRefCode()
	if err != nil {
		t.Fatalf("สุ่ม Ref Code พัง: %v", err)
	}

	if len(ref) != 6 {
		t.Errorf("Ref Code ต้องมี 6 ตัว แต่ได้มา: %s", ref)
	}
}

// 3. ทดสอบการ Hash Citizen ID (ต้องได้ค่าเดิมเสมอถ้า Key เดิม)
func TestGenerateCitizenIDHash(t *testing.T) {
	// ตั้งค่า Secret ชั่วคราวสำหรับการเทสต์
	os.Setenv("HASH_SECRET_KEY", "test_secret_key")

	id := "1234567890123"
	hash1 := generateCitizenIDHash(id)
	hash2 := generateCitizenIDHash(id)

	if hash1 == "" {
		t.Error("Hash ที่ได้ต้องไม่ว่างเปล่า")
	}

	if hash1 != hash2 {
		t.Error("Hash ค่าเดิมต้องได้ผลลัพธ์เดิม (Idempotent)")
	}
}

// 4. ทดสอบกรณีที่ Secret Key เปลี่ยนไป (Hash ต้องเปลี่ยน)
func TestGenerateCitizenIDHash_DifferentKey(t *testing.T) {
	id := "1234567890123"

	os.Setenv("HASH_SECRET_KEY", "key_1")
	hash1 := generateCitizenIDHash(id)

	os.Setenv("HASH_SECRET_KEY", "key_2")
	hash2 := generateCitizenIDHash(id)

	if hash1 == hash2 {
		t.Error("Secret Key ต่างกัน Hash ไม่ควรจะเหมือนกันนะ!")
	}
}
