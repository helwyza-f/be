package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

// CheckMe digunakan frontend untuk memastikan token masih valid saat page reload
func (h *Handler) CheckMe(c *gin.Context) {
	userID, _ := c.Get("userID")
	tenantID, _ := c.Get("tenantID")
	role, _ := c.Get("role")

	c.JSON(http.StatusOK, gin.H{
		"user_id":   userID,
		"tenant_id": tenantID,
		"role":      role,
		"status":    "authenticated",
	})
}