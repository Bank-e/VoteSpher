package models

import "time"

type Vote struct {
	VoteID      uint       `gorm:"primaryKey;autoIncrement"`
	AreaID      uint       `gorm:"not null"`
	Area        Area       `gorm:"foreignKey:AreaID"`
	CandidateID *uint      // nullable = Vote No เขต
	Candidate   *Candidate `gorm:"foreignKey:CandidateID"`
	PartyID     *uint      // nullable = Vote No พรรค
	Party       *Party     `gorm:"foreignKey:PartyID"`
	CreatedAt   time.Time
}