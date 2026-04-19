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
