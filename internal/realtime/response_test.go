package realtime

import "testing"

func TestBuildResponse(t *testing.T) {
	areaRows := []AreaVoteRow{
		{AreaID: 1, AreaName: "A", TotalVotes: 100},
		{AreaID: 2, AreaName: "B", TotalVotes: 200},
	}
	candidateRows := []AreaCandidateRow{
		{AreaID: 1, CandidateNo: 1, CandidateName: "Alice", PartyName: "Party X", Votes: 60},
		{AreaID: 1, CandidateNo: 2, CandidateName: "Bob", PartyName: "Party Y", Votes: 40},
		{AreaID: 2, CandidateNo: 3, CandidateName: "Charlie", PartyName: "Party X", Votes: 120},
	}
	partyRows := []PartyVoteRow{
		{PartyNo: 1, PartyName: "Party X", Votes: 180},
		{PartyNo: 2, PartyName: "Party Y", Votes: 120},
	}
	resp := buildResponse(areaRows, candidateRows, partyRows)
	if resp.TotalVotes != 300 {
		t.Errorf("expected 300, got %d", resp.TotalVotes)
	}
	if len(resp.Areas) != 2 {
		t.Errorf("expected 2 areas, got %d", len(resp.Areas))
	}
	if len(resp.Areas[0].Candidates) != 2 {
		t.Errorf("expected 2 candidates in area 1, got %d", len(resp.Areas[0].Candidates))
	}
	if len(resp.Areas[1].Candidates) != 1 {
		t.Errorf("expected 1 candidate in area 2, got %d", len(resp.Areas[1].Candidates))
	}
	if len(resp.Party) != 2 {
		t.Errorf("expected 2 parties, got %d", len(resp.Party))
	}
	if resp.Party[0].PartyName != "Party X" {
		t.Errorf("expected first party 'Party X', got '%s'", resp.Party[0].PartyName)
	}
	if resp.LastUpdated == "" {
		t.Error("expected last_updated to be set")
	}
}
