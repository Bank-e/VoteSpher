package models

type Party struct {
	PartyID   uint   `gorm:"primaryKey;autoIncrement"`
	PartyNo   int    `gorm:"not null;uniqueIndex"`
	PartyName string `gorm:"type:varchar(255);not null"`
	LogoURL   string `gorm:"type:varchar(500)"`
}