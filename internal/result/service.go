package result

type ResultService interface {
	GetAreaResult(areaID uint) (AreaResultResponse, error)
}

type resultService struct {
	repo ResultRepository
}

func NewResultService(repo ResultRepository) ResultService {
	return &resultService{repo: repo}
}

func (s *resultService) GetAreaResult(areaID uint) (AreaResultResponse, error) {
	return s.repo.GetVoteResultByArea(areaID)
}
