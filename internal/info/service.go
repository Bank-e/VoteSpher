package info

import "gorm.io/gorm"

func GetCandidatesService(db *gorm.DB, areaID string) ([]Candidate, error) {
	return GetCandidates(db, areaID)
}

func GetPartiesService(db *gorm.DB) ([]Party, error) {
	return GetParties(db)
}
