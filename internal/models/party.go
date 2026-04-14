package models

type Party struct {
    ID        uint   `gorm:"primaryKey;autoIncrement;column:party_id"`
    PartyNo   int    `gorm:"not null;uniqueIndex"`
    PartyName string `gorm:"type:varchar(255);not null"`
    LogoURL   string `gorm:"type:varchar(500)"`
}