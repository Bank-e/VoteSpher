package realtime

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetVoteResultByAreaHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		areaID := c.Param("area_id")
		if _, err := strconv.Atoi(areaID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "area_id ต้องเป็นตัวเลข"})
			return
		}
		results, err := GetAreaVotes(db, areaID)
		if err != nil {
			log.Printf("GetVoteResultByArea error area=%s: %v", areaID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "เกิดข้อผิดพลาดภายในระบบ"})
			return
		}
		c.JSON(http.StatusOK, results)
	}
}

func GetAllAreasVotesHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := GetAllAreasVotes(db)
		if err != nil {
			log.Printf("GetAllAreasVotes error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "เกิดข้อผิดพลาดภายในระบบ"})
			return
		}
		c.JSON(http.StatusOK, BuildResponse(rows))
	}
}
