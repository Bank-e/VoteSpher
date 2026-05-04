package result

import (
	"votespher/internal/models"

	"gorm.io/gorm"
)

type ResultRepository interface {
	GetVoteResultByArea(areaID uint) (AreaResultResponse, error)
}

type resultRepository struct {
	db *gorm.DB
}

func NewResultRepository(db *gorm.DB) ResultRepository {
	return &resultRepository{db: db}
}

func (r *resultRepository) GetVoteResultByArea(areaID uint) (AreaResultResponse, error) {
	var area models.Area

	if err := r.db.Where("area_id = ?", areaID).First(&area).Error; err != nil {
		return AreaResultResponse{}, err
	}

	var candidateResults []CandidateResult
	if err := r.db.
		Table("votes").
		Select("candidates.candidate_no, candidates.full_name AS name, COUNT(votes.vote_id) AS votes").
		Joins("JOIN candidates ON votes.candidate_id = candidates.candidate_id").
		Where("votes.area_id = ?", areaID).
		Group("candidates.candidate_no, candidates.full_name").
		Order("votes DESC").
		Scan(&candidateResults).Error; err != nil {
		return AreaResultResponse{}, err
	}

	var partyResults []PartyResult
	if err := r.db.
		Table("votes").
		Select("parties.party_no, parties.party_name, COUNT(votes.vote_id) AS votes").
		Joins("JOIN parties ON votes.party_id = parties.party_id").
		Where("votes.area_id = ?", areaID).
		Group("parties.party_no, parties.party_name").
		Order("votes DESC").
		Scan(&partyResults).Error; err != nil {
		return AreaResultResponse{}, err
	}

	var lastUpdated string
	_ = r.db.
		Table("votes").
		Select("MAX(created_at)").
		Where("area_id = ?", areaID).
		Scan(&lastUpdated).Error

	return AreaResultResponse{
		AreaID:           areaID,
		AreaName:         area.AreaName,
		LastUpdated:      lastUpdated,
		CandidateResults: candidateResults,
		PartyListResults: partyResults,
	}, nil
}
