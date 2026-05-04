package result

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAreaResultHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		areaIDParam := c.Param("id")

		areaID, err := strconv.Atoi(areaIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid id: must be a number",
			})
			return
		}

		result, err := GetAreaResultService(db, uint(areaID))
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
