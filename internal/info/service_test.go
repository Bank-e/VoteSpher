package info

import "testing"

// mock repository
type mockRepo struct{}

func (m *mockRepo) GetCandidates(areaID int) ([]Candidate, error) {
	return []Candidate{
		{Name: "Mock Candidate"},
	}, nil
}

func (m *mockRepo) GetParties() ([]Party, error) {
	return []Party{
		{PartyName: "Mock Party"},
	}, nil
}

func TestServiceGetCandidates(t *testing.T) {
	repo := &mockRepo{}
	service := NewInfoService(repo)

	result, err := service.GetCandidates(1)

	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
}

func TestServiceGetParties(t *testing.T) {
	repo := &mockRepo{}
	service := NewInfoService(repo)

	result, err := service.GetParties()

	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
}
