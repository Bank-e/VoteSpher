package models

import "time"

type OTP struct {
	OTPID     uint      `gorm:"primaryKey;autoIncrement"`
	VoterID   uint      `gorm:"not null"`
	Voter     Voter     `gorm:"foreignKey:VoterID"`
	OTPCode   string    `gorm:"type:varchar(10);not null"`
	RefCode   string    `gorm:"type:varchar(10);not null"`
	ExpiresAt time.Time `gorm:"not null"`
	IsUsed    bool      `gorm:"default:false"`
}