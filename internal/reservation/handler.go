package reservation

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateBookingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "LENGKAPI DATA BOOKING"})
		return
	}

	// Logic tenant detection
	if req.TenantID == "" {
		if tID, exists := c.Get("tenantID"); exists {
			req.TenantID = tID.(string)
		}
	}

	b, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "BOOKING BERHASIL",
		"booking_id":   b.ID,
		"access_token": b.AccessToken,
		"redirect_url": "/status/" + b.AccessToken.String(),
	})
}

func (h *Handler) Availability(c *gin.Context) {
	resourceID := c.Param("resource_id")
	dateStr := c.Query("date")

	targetDate := time.Now()
	if dateStr != "" {
		if d, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = d
		}
	}

	busy, err := h.service.GetAvailability(c.Request.Context(), resourceID, targetDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"busy_slots": busy,
	})
}

func (h *Handler) Status(c *gin.Context) {
	token := c.Param("token")
	b, err := h.service.GetStatusByToken(c.Request.Context(), token)
	if err != nil || b == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Detail booking tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, b)
}

// ListAll untuk Admin Dashboard
func (h *Handler) ListAll(c *gin.Context) {
    tenantID := c.MustGet("tenantID").(string)
    status := c.Query("status") // Optional filter: ?status=pending

    bookings, err := h.service.ListByTenant(c.Request.Context(), tenantID, status)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, bookings)
}

// GetDetail untuk Admin melihat rincian (termasuk items/options)
func (h *Handler) GetDetail(c *gin.Context) {
    id := c.Param("id")
    tenantID := c.MustGet("tenantID").(string)

    booking, err := h.service.GetDetailForAdmin(c.Request.Context(), id, tenantID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Booking tidak ditemukan"})
        return
    }
    c.JSON(http.StatusOK, booking)
}

// UpdateStatus (Confirm, Ongoing, Cancelled, Finished)
func (h *Handler) UpdateStatus(c *gin.Context) {
    id := c.Param("id")
    tenantID := c.MustGet("tenantID").(string)
    
    var req struct {
        Status string `json:"status" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Status wajib diisi"})
        return
    }

    err := h.service.UpdateStatus(c.Request.Context(), id, tenantID, req.Status)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Status booking berhasil diperbarui ke " + req.Status})
}

