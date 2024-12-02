package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"google.golang.org/api/idtoken"
	"net/http"
	"os"
	"strings"
	"time"
)

func CdnCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			c.Header("Cache-Control", "public, max-age=360, s-maxage=31536000")
			c.Header("Expires", time.Now().AddDate(10, 0, 0).Format(http.TimeFormat)) // 10 years in the future
		}
		c.Next()
	}
}

func CorsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowOrigins:     strings.Split(os.Getenv("ALLOWED_ORIGINS"), ","),
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, // How long the results of a preflight request can be cached
	})
}

func OtelGinMiddleware() gin.HandlerFunc {
	return otelgin.Middleware(os.Getenv("PROD_DOMAIN"))
}

func GcpPubSubMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if os.Getenv("ENVIRONMENT") == "dev" {
			c.Next()
			return
		}
		// Validate the GCP Pub/Sub request
		// Verify the ID token
		_, err := idtoken.Validate(c.Request.Context(), c.Request.Header.Get("Authorization"), os.Getenv("GCP_PROJECT_ID"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid ID token"})
			c.Abort()
			return
		}
		c.Next()
	}
}
