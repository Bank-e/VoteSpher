package realtime

type CandidateResponse struct {
	CandidateID   int    `json:"candidate_id"`
	CandidateName string `json:"candidate_name"`
	Votes         int    `json:"votes"`
}

type AreaResponse struct {
	AreaID     int                 `json:"area_id"`
	AreaName   string              `json:"area_name"`
	TotalVotes int                 `json:"total_votes"`
	Candidates []CandidateResponse `json:"candidates"`
}

type Response struct {
	TotalVotes  int            `json:"total_votes"`
	LastUpdated string         `json:"last_updated"`
	Areas       []AreaResponse `json:"areas"`
}
