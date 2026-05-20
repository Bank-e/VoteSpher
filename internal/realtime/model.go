package realtime

// ===== Row types for DB queries =====

type AreaVoteRow struct {
	AreaID     int
	AreaName   string
	TotalVotes int
}

type AreaCandidateRow struct {
	AreaID        int
	CandidateNo   int
	CandidateName string
	PartyName     string
	Votes         int
}

type PartyVoteRow struct {
	PartyNo   int
	PartyName string
	Votes     int
}
