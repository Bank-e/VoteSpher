package result

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetProvinceAreaResultHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		provinceName := c.Param("provinces_name")
		areaID := c.Param("area_id")

		fmt.Println(provinceName, areaID)

		_, err := strconv.Atoi(areaID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid area_id: must be a number",
			})
			return
		}

		result, err := GetProvinceAreaResultService(db, provinceName, areaID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "area not found",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
