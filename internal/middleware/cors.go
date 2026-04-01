package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Logika Dinamis: Izinkan jika origin mengandung bookinaja.com atau localhost
		// Ini biar subdomain tenant (miniboss.bookinaja.com) nggak kena block
		if origin != "" && (strings.HasSuffix(origin, "bookinaja.com") || strings.Contains(origin, "localhost")) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// Fallback jika tidak cocok, tetap set origin agar tidak error di browser tertentu
			// Tapi untuk tahap dev/awal, kamu bisa langsung set:
			// c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		// Handle Preflight Request (Sangat Penting!)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
