package models

type Candidate struct {
    ID          uint   `gorm:"primaryKey;autoIncrement;column:candidate_id"`
    AreaID      uint   `gorm:"not null"`
    Area        Area   
    PartyID     uint   `gorm:"not null"`
    Party       Party  
    CandidateNo int    `gorm:"not null"`
    FullName    string `gorm:"type:varchar(255);not null"`
    Biography   string `gorm:"type:text"`
}