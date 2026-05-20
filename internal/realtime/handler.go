package realtime

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RealtimeHandler struct {
	service RealtimeService
}

func NewRealtimeHandler(service RealtimeService) *RealtimeHandler {
	return &RealtimeHandler{service: service}
}

// GET /results/areas
func (h *RealtimeHandler) GetAllAreasVotes(c *gin.Context) {
	result, err := h.service.GetAllAreasResult()
	if err != nil {
		log.Printf("GetAllAreasVotes error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "เกิดข้อผิดพลาดภายในระบบ"})
		return
	}
	c.JSON(http.StatusOK, result)
}
