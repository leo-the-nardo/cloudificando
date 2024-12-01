package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
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
