package realtime

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetVoteResultByAreaHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		areaID := c.Param("area_id")

		results, err := GetAreaVotes(db, areaID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, results)
	}
}
func GetAllAreasVotesHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		rows, err := GetAllAreasVotes(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		response := BuildResponse(rows)

		c.JSON(http.StatusOK, response)
	}
}
