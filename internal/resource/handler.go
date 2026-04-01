package resource

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/helwiza/saas/internal/booking"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Create(c *gin.Context) {
	tenantID := c.MustGet("tenantID").(string)
	var req struct {
		Name     string `json:"name" binding:"required"`
		Category string `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.service.CreateResource(c.Request.Context(), tenantID, req.Name, req.Category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *Handler) List(c *gin.Context) {
	tenantID := c.MustGet("tenantID").(string)
	resources, err := h.service.ListResources(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resources)
}

func (h *Handler) AddItem(c *gin.Context) {
    resourceID := c.Param("id")
    var req struct {
        Name         string  `json:"name" binding:"required"`
        PricePerHour float64 `json:"price_per_hour"`
        ItemType     string  `json:"item_type" binding:"required"` // 'main' atau 'addon'
        IsDefault    bool    `json:"is_default"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Lengkapi data barang: " + err.Error()})
        return
    }

    // Panggil service dengan parameter lengkap
    item, err := h.service.AddResourceItem(c.Request.Context(), resourceID, req.Name, req.PricePerHour, req.ItemType, req.IsDefault)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, item)
}

func (h *Handler) ListItems(c *gin.Context) {
	id := c.Param("id")
	items, err := h.service.GetItems(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteResource(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Resource deleted successfully"})
}

// UpdateItem menangani PUT /api/v1/resources-all/items/:id
func (h *Handler) UpdateItem(c *gin.Context) {
	id := c.Param("id")
	var req booking.ResourceItem
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateItem(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item updated successfully"})
}

func (h *Handler) DeleteItem(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteItem(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Item deleted successfully"})
}

func (h *Handler) GetPublicDetail(c *gin.Context) {
    id := c.Param("id")
    res, err := h.service.GetResourceDetail(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Resource tidak ditemukan"})
        return
    }
    c.JSON(http.StatusOK, res)
}