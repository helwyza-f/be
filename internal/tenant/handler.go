package tenant

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/helwiza/saas/internal/platform/storage"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) GetPublicLandingData(c *gin.Context) {
	slug := c.Query("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
		return
	}

	cleanSlug := strings.Split(slug, ".")[0]

	tenant, err := h.service.repo.GetBySlug(c.Request.Context(), cleanSlug)
	if err != nil || tenant == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bisnis tidak ditemukan"})
		return
	}

	// Ambil resources yang ID-nya sudah diconvert ke text dan items sudah di-unmarshal
	resources, err := h.service.repo.ListResourcesWithItems(c.Request.Context(), tenant.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data unit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"profile":   tenant,
		"resources": resources,
	})
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"tenant":    t,
		"login_url": fmt.Sprintf("https://%s.%s", t.Slug, "bookinaja.com"),
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetProfile(c *gin.Context) {
	tIDRaw, _ := c.Get("tenantID")
	tID, _ := uuid.Parse(tIDRaw.(string))
	p, err := h.service.GetProfile(c.Request.Context(), tID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	tIDRaw, _ := c.Get("tenantID")
	tID, _ := uuid.Parse(tIDRaw.(string))

	var req Tenant
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.service.UpdateProfile(c.Request.Context(), tID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *Handler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada gambar yang diupload"})
		return
	}
	tenantID := c.MustGet("tenantID").(string)
	s3Provider, err := storage.NewS3Client()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Storage error"})
		return
	}
	url, err := s3Provider.UploadFile(c.Request.Context(), file, "tenants/"+tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}
