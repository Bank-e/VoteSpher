package models

import "time"

type SystemConfig struct {
    ID        uint      `gorm:"primaryKey;autoIncrement;column:config_id"`
    AdminID   uint      `gorm:"not null"`
    Admin     Admin     
    Status    string    `gorm:"type:varchar(50);not null"`
    StartTime time.Time `gorm:"not null"`
    EndTime   time.Time `gorm:"not null"`
    UpdatedAt time.Time
    IsActive  bool      `gorm:"default:false"`
}