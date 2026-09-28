package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"nexus-mesh-go/internal/store"
)

func HashAPIKey(apiKey string) string {
	h := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(h[:])
}

func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"detail": "Missing API key in header. Provide 'X-API-Key' or 'Authorization: Bearer <key>'.",
			})
			c.Abort()
			return
		}

		keyHash := HashAPIKey(apiKey)
		var merchant store.Merchant
		if err := db.Where("api_key_hash = ?", keyHash).First(&merchant).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"detail": "Invalid or revoked API key.",
			})
			c.Abort()
			return
		}

		c.Set("merchant", &merchant)
		c.Next()
	}
}
