package info

// 🔹 Interface
type InfoService interface {
	GetCandidates(areaID int) ([]Candidate, error)
	GetParties() ([]Party, error)
}

// 🔹 Struct
type infoService struct {
	repo InfoRepository
}

// 🔹 Constructor
func NewInfoService(repo InfoRepository) InfoService {
	return &infoService{repo: repo}
}

// 🔹 Implementation
func (s *infoService) GetCandidates(areaID int) ([]Candidate, error) {
	return s.repo.GetCandidates(areaID)
}

func (s *infoService) GetParties() ([]Party, error) {
	return s.repo.GetParties()
}
