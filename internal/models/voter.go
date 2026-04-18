package models

import "time"

type Voter struct {
    ID            uint       `gorm:"primaryKey;autoIncrement;column:voter_id"`
    CitizenIDHash string     `gorm:"type:varchar(255);not null;uniqueIndex"`
    AreaID        uint       `gorm:"not null"`
    Area          Area       
    PhoneNumber   string     `gorm:"type:varchar(20);not null"`
    IsVoted       bool       `gorm:"default:false"`
    VotedAt       *time.Time
}