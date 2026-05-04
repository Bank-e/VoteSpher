package info

import (
	"context"

	"gorm.io/gorm"
)

// 🔹 Interface
type InfoRepository interface {
	GetCandidates(areaID int) ([]Candidate, error)
	GetParties() ([]Party, error)
}

// 🔹 Struct
type infoRepository struct {
	db *gorm.DB
}

// 🔹 Constructor
func NewInfoRepository(db *gorm.DB) InfoRepository {
	return &infoRepository{db: db}
}

// 🔹 Implementation
func (r *infoRepository) GetCandidates(areaID int) ([]Candidate, error) {
	var result []Candidate

	err := r.db.WithContext(context.Background()).
		Table("candidates").
		Select(`
			candidates.candidate_no,
			candidates.full_name as name,
			candidates.party_id,
			parties.party_name,
			parties.logo_url,
			candidates.biography
		`).
		Joins("JOIN parties ON candidates.party_id = parties.party_id").
		Where("candidates.area_id = ?", areaID).
		Scan(&result).Error

	return result, err
}

func (r *infoRepository) GetParties() ([]Party, error) {
	var result []Party

	err := r.db.Table("parties").
		Select(`
			party_id,
			party_no,
			party_name,
			logo_url
		`).
		Scan(&result).Error

	return result, err
}
