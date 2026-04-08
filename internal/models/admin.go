package models

type Admin struct {
	AdminID uint  `gorm:"primaryKey;autoIncrement"`
	VoterID uint  `gorm:"not null;uniqueIndex"`
	Voter   Voter `gorm:"foreignKey:VoterID"`
}