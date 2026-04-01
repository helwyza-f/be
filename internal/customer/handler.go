package customer

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/helwiza/saas/internal/platform/fonnte"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

// ValidatePhone digunakan untuk live validation di Frontend saat admin mengetik nomor
func (h *Handler) ValidatePhone(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nomor WhatsApp wajib diisi"})
		return
	}

	// Live check via Fonnte API
	isValid, err := fonnte.ValidateNumber(phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"valid": false, 
			"error": "Gagal terhubung ke server WhatsApp",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": isValid,
		"phone": phone,
	})
}

func (h *Handler) Create(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data pelanggan salah"})
		return
	}

	tenantID := c.MustGet("tenantID").(string)
	cust, err := h.service.Register(c.Request.Context(), tenantID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, cust)
}

func (h *Handler) List(c *gin.Context) {
	tenantID := c.MustGet("tenantID").(string)
	customers, err := h.service.ListByTenant(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, customers)
}