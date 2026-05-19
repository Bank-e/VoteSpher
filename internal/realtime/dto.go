package realtime

type CandidateResponse struct {
	CandidateNo   int    `json:"candidate_no"`
	CandidateName string `json:"candidate_name"`
	PartyName     string `json:"party_name"`
	Votes         int    `json:"votes"`
}

type AreaResponse struct {
	AreaID     int                 `json:"area_id"`
	AreaName   string              `json:"area_name"`
	TotalVotes int                 `json:"total_votes"`
	Candidates []CandidateResponse `json:"candidates"`
}

type PartyResponse struct {
	PartyNo   int    `json:"party_no"`
	PartyName string `json:"party_name"`
	Votes     int    `json:"votes"`
}

type Response struct {
	TotalVotes  int             `json:"total_votes"`
	LastUpdated string          `json:"last_updated"`
	Areas       []AreaResponse  `json:"areas"`
	Party       []PartyResponse `json:"party"`
}
