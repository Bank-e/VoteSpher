package result

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetProvinceAreaResultHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		provinceName := c.Param("provinces_name")
		areaID := c.Param("area_id")
		fmt.Println(provinceName, areaID)

		result, err := GetProvinceAreaResultService(db, provinceName, areaID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
