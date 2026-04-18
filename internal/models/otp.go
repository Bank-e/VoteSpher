package models

import "time"

type OTP struct {
    ID        uint      `gorm:"primaryKey;autoIncrement;column:otp_id"`
    VoterID   uint      `gorm:"not null"`
    Voter     Voter     
    OTPCode   string    `gorm:"type:varchar(10);not null"`
    RefCode   string    `gorm:"type:varchar(10);not null"`
    ExpiresAt time.Time `gorm:"not null"`
    IsUsed    bool      `gorm:"default:false"`
}