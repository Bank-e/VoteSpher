package info

type Candidate struct {
	CandidateNo int    `json:"candidate_no"`
	Name        string `json:"name"`
	PartyID     int    `json:"party_id"`
	PartyName   string `json:"party_name"`
	LogoURL     string `json:"logo_url"`
	Biography   string `json:"biography"`
}

type Party struct {
	PartyID   int    `json:"party_id"`
	PartyNo   int    `json:"party_no"`
	PartyName string `json:"party_name"`
	LogoURL   string `json:"logo_url"`
}