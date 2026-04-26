package info

type Candidate struct {
	CandidateNo int    `gorm:"column:candidate_no"`
	Name        string `gorm:"column:name"` 
	PartyID     int    `gorm:"column:party_id"`
	PartyName   string `gorm:"column:party_name"`
	LogoURL     string `gorm:"column:logo_url"`
	Biography   string `gorm:"column:biography"`
}

type Party struct {
	PartyID   int    `json:"party_id"`
	PartyNo   int    `json:"party_no"`
	PartyName string `json:"party_name"`
	LogoURL   string `json:"logo_url"`
}