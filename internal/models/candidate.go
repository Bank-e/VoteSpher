package models

type Candidate struct {
	CandidateID  uint   `gorm:"primaryKey;autoIncrement"`
	AreaID       uint   `gorm:"not null"`
	Area         Area   `gorm:"foreignKey:AreaID"`
	PartyID      uint   `gorm:"not null"`
	Party        Party  `gorm:"foreignKey:PartyID"`
	CandidateNo  int    `gorm:"not null"`
	FullName     string `gorm:"type:varchar(255);not null"`
	Biography    string `gorm:"type:text"`
}