package result

type AreaResultResponse struct {
	AreaID           uint              `json:"area_id"`
	AreaName         string            `json:"area_name"`
	LastUpdated      string            `json:"last_updated"`
	CandidateResults []CandidateResult `json:"candidate_results"`
	PartyListResults []PartyResult     `json:"party_list_results"`
}

type CandidateResult struct {
	CandidateNo int    `json:"candidate_no"`
	Name        string `json:"name"`
	Votes       int    `json:"votes"`
}

type PartyResult struct {
	PartyNo   int    `json:"party_no"`
	PartyName string `json:"party_name"`
	Votes     int    `json:"votes"`
}
