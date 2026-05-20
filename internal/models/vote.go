package models

import "time"

type Vote struct {
    ID          uint       `gorm:"primaryKey;autoIncrement;column:vote_id"`
    AreaID      uint       `gorm:"not null"`
    Area        Area       
    CandidateID *uint      
    Candidate   *Candidate 
    PartyID     *uint      
    Party       *Party     
    CreatedAt   time.Time
}