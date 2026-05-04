package result

import (
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

func (h *ResultHandler) GetAreaResult(c *gin.Context) {
	areaIDParam := c.Param("id")

	areaID, err := strconv.Atoi(areaIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id: must be a number",
		})
		return
	}

	result, err := h.service.GetAreaResult(uint(areaID))
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
