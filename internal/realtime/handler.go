package realtime

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type RealtimeHandler struct {
	service RealtimeService
}

func NewRealtimeHandler(service RealtimeService) *RealtimeHandler {
	return &RealtimeHandler{service: service}
}

func (h *RealtimeHandler) GetAllAreasVotes(c *gin.Context) {

	result, err := h.service.GetAllAreasResult()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
