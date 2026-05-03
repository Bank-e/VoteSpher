package pkg

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTCustomClaims struct {
	VoterID uint   `json:"voter_id"`
	AreaID  uint   `json:"area_id"`
	Role    string `json:"role"` // "voter" หรือ "admin"
	jwt.RegisteredClaims         // ตัวนี้จะจัดการเรื่อง วันหมดอายุ (exp), วันที่ออก (iat) ให้อัตโนมัติ
}

// ฟังก์ชันสร้าง Token (ใช้ตอนยืนยัน OTP สำเร็จ)
func GenerateToken(voterID uint, areaID uint, role string, secretKey string) (string, error) {
	// กำหนดอายุของ Token (เช่น ให้มีอายุ 2 ชั่วโมง)
	expirationTime := time.Now().Add(2 * time.Hour)

	// นำข้อมูลมาใส่ใน Claims
	claims := &JWTCustomClaims{
		VoterID: voterID,
		AreaID:  areaID,
		Role:    role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "VoteSpher", // ระบุชื่อระบบที่ออก Token
		},
	}

	// สร้าง Token ด้วย Algorithm HS256 พร้อมยัด Claims ลงไป
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// เซ็นรับรอง (Sign) ด้วย Secret Key และแปลงเป็น String เพื่อส่งให้ Frontend
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ฟังก์ชันอ่านและตรวจสอบ Token (ใช้ตอนดึงข้อมูลหรือทำ Middleware)
func ValidateToken(tokenString string, secretKey string) (*JWTCustomClaims, error) {
	// Parse Token พร้อมถอดรหัสออกมาใส่ Struct ที่เราเตรียมไว้
	token, err := jwt.ParseWithClaims(tokenString, &JWTCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// ตรวจสอบว่า Algorithm ที่ใช้เซ็นตรงกับที่เราตั้งไว้หรือไม่ (ป้องกันการปลอมแปลง)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// คืนค่า Secret Key เพื่อให้ระบบใช้แกะข้อมูล
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err // อาจจะหมดอายุ (Token is expired) หรือ Token ผิดรูปแบบ
	}

	// แกะข้อมูลออกมาใช้งาน หาก Token ถูกต้องและข้อมูลเข้าข่าย Custom Claims ของเรา
	if claims, ok := token.Claims.(*JWTCustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}