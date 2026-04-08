package models

import "time"

type SystemConfig struct {
	ConfigID  uint      `gorm:"primaryKey;autoIncrement"`
	AdminID   uint      `gorm:"not null"`
	Admin     Admin     `gorm:"foreignKey:AdminID"`
	Status    string    `gorm:"type:varchar(50);not null"`
	StartTime time.Time `gorm:"not null"`
	EndTime   time.Time `gorm:"not null"`
	UpdatedAt time.Time
	IsActive  bool      `gorm:"default:false"`
}