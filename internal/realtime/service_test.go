package realtime

import (
	"errors"
	"testing"
)

type mockRealtimeRepo struct {
	areaRows      []AreaVoteRow
	areaErr       error
	candidateRows []AreaCandidateRow
	candidateErr  error
	partyRows     []PartyVoteRow
	partyErr      error
}

func (m *mockRealtimeRepo) GetAllAreasVotes() ([]AreaVoteRow, error) {
	return m.areaRows, m.areaErr
}
func (m *mockRealtimeRepo) GetTopCandidatesByArea(limit int) ([]AreaCandidateRow, error) {
	return m.candidateRows, m.candidateErr
}
func (m *mockRealtimeRepo) GetPartyVotes() ([]PartyVoteRow, error) {
	return m.partyRows, m.partyErr
}

func TestGetAllAreasResult_Success(t *testing.T) {
	repo := &mockRealtimeRepo{
		areaRows:      []AreaVoteRow{{AreaID: 1, AreaName: "Area A", TotalVotes: 10}},
		candidateRows: []AreaCandidateRow{{AreaID: 1, CandidateNo: 1, CandidateName: "Alice", Votes: 10}},
		partyRows:     []PartyVoteRow{{PartyNo: 1, PartyName: "Party X", Votes: 10}},
	}
	svc := NewRealtimeService(repo)
	resp, err := svc.GetAllAreasResult()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalVotes != 10 {
		t.Errorf("expected TotalVotes=10, got %d", resp.TotalVotes)
	}
}

func TestGetAllAreasResult_GetAllAreasVotesError(t *testing.T) {
	repo := &mockRealtimeRepo{areaErr: errors.New("db error")}
	svc := NewRealtimeService(repo)
	if _, err := svc.GetAllAreasResult(); err == nil {
		t.Error("expected error from GetAllAreasVotes")
	}
}

func TestGetAllAreasResult_GetTopCandidatesError(t *testing.T) {
	repo := &mockRealtimeRepo{
		areaRows:     []AreaVoteRow{},
		candidateErr: errors.New("db error"),
	}
	svc := NewRealtimeService(repo)
	if _, err := svc.GetAllAreasResult(); err == nil {
		t.Error("expected error from GetTopCandidatesByArea")
	}
}

func TestGetAllAreasResult_GetPartyVotesError(t *testing.T) {
	repo := &mockRealtimeRepo{
		areaRows:      []AreaVoteRow{},
		candidateRows: []AreaCandidateRow{},
		partyErr:      errors.New("db error"),
	}
	svc := NewRealtimeService(repo)
	if _, err := svc.GetAllAreasResult(); err == nil {
		t.Error("expected error from GetPartyVotes")
	}
}

func TestBuildResponse_EmptyInputs(t *testing.T) {
	resp := buildResponse(nil, nil, nil)
	if resp.TotalVotes != 0 {
		t.Errorf("expected 0, got %d", resp.TotalVotes)
	}
	if len(resp.Areas) != 0 {
		t.Errorf("expected empty areas")
	}
	if len(resp.Party) != 0 {
		t.Errorf("expected empty party")
	}
	if resp.LastUpdated == "" {
		t.Error("expected last_updated to be set")
	}
}
