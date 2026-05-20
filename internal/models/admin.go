package models

type Admin struct {
	ID      uint  `gorm:"primaryKey;autoIncrement;column:admin_id"`
	VoterID uint  `gorm:"not null;uniqueIndex"`
	Voter   Voter
}