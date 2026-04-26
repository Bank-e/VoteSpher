package info

import "gorm.io/gorm"

func GetCandidates(db *gorm.DB, areaID int) ([]Candidate, error) {
	var result []Candidate

	err := db.Table("candidates").
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

func GetParties(db *gorm.DB) ([]Party, error) {
	var result []Party

	err := db.Table("parties").
		Select(`
			party_id,
			party_no,
			party_name,
			logo_url
		`).
		Scan(&result).Error

	return result, err
}