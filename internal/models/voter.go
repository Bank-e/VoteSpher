package models

import "time"

type Voter struct {
	VoterID        uint      `gorm:"primaryKey;autoIncrement"`
	CitizenIDHash  string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	AreaID         uint      `gorm:"not null"`
	Area           Area      `gorm:"foreignKey:AreaID"`
	PhoneNumber    string    `gorm:"type:varchar(20);not null"`
	IsVoted        bool      `gorm:"default:false"`
	VotedAt        *time.Time
}