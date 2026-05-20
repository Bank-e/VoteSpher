package info

type Candidate struct {
	CandidateID int    `gorm:"column:candidate_id" json:"candidate_id"`
	CandidateNo int    `gorm:"column:candidate_no" json:"candidate_no"`
	Name        string `gorm:"column:name"         json:"name"`
	PartyID     int    `gorm:"column:party_id"     json:"party_id"`
	PartyName   string `gorm:"column:party_name"   json:"party_name"`
	LogoURL     string `gorm:"column:logo_url"     json:"logo_url"`
	Biography   string `gorm:"column:biography"    json:"biography"`
}

type Party struct {
	PartyID   int    `json:"party_id"`
	PartyNo   int    `json:"party_no"`
	PartyName string `json:"party_name"`
	LogoURL   string `json:"logo_url"`
}
