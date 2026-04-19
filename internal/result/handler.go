package result

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetProvinceAreaResultHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		provinceName := c.Param("provinces_name")
		areaID := c.Param("area_id")

		c.JSON(http.StatusOK, AreaResultResponse{
			ProvinceName: provinceName,
			AreaID:       areaID,
			Message:      "API is working!",
		})
	}
}
