package result

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ResultHandler struct {
	service ResultService
}

func NewResultHandler(service ResultService) *ResultHandler {
	return &ResultHandler{service: service}
}

// GET /results/areas/:area_id
func (h *ResultHandler) GetAreaResult(c *gin.Context) {
	areaIDParam := c.Param("area_id")
	areaID, err := strconv.Atoi(areaIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid area_id: must be a number"})
		return
	}

	result, err := h.service.GetAreaResult(uint(areaID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "area not found"})
			return
		}
		log.Printf("GetAreaResult error area=%d: %v", areaID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "เกิดข้อผิดพลาดภายในระบบ"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GET /results/provinces/:provinces_name/areas/:area_id
func (h *ResultHandler) GetProvinceAreaResult(c *gin.Context) {
	areaIDParam := c.Param("area_id")
	areaID, err := strconv.Atoi(areaIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid area_id: must be a number"})
		return
	}

	result, err := h.service.GetAreaResult(uint(areaID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "area not found"})
			return
		}
		log.Printf("GetProvinceAreaResult error area=%d: %v", areaID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "เกิดข้อผิดพลาดภายในระบบ"})
		return
	}

	c.JSON(http.StatusOK, result)
}
