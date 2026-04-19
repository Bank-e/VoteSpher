package result

import (
	"votespher/internal/models"

	"gorm.io/gorm"
)

func GetProvinceAreaResultRepository(db *gorm.DB, provinceName string, areaID string) (AreaResultResponse, error) {
	var area models.Area

	err := db.Where("area_id = ?", areaID).First(&area).Error
	if err != nil {
		return AreaResultResponse{}, err
	}

	return AreaResultResponse{
		ProvinceName: provinceName, // ยัง mock ไว้ก่อน
		AreaID:       areaID,
		Message:      area.AreaName, // ใช้ของจริงจาก DB
	}, nil
}

func GetVoteResultByArea(db *gorm.DB, areaID string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	err := db.
		Table("votes").
		Select("parties.party_name, COUNT(*) as total").
		Joins("JOIN parties ON votes.party_id = parties.party_id").
		Where("votes.area_id = ?", areaID).
		Group("parties.party_name").
		Find(&results).Error

	return results, err
}
