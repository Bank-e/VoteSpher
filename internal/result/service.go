package result

import "gorm.io/gorm"

func GetProvinceAreaResultService(db *gorm.DB, provinceName string, areaID string) (AreaResultResponse, error) {
	return GetProvinceAreaResultRepository(db, provinceName, areaID)
}
