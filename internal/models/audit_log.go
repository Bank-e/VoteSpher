package models

import "time"

type AuditLog struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	VoterID   *uint     `gorm:"index"`
	Action    string    `gorm:"type:varchar(100);not null"`
	Detail    string    `gorm:"type:text"`
	IPAddress string    `gorm:"type:varchar(45)"`
	CreatedAt time.Time
}
