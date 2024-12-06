package main

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"google.golang.org/api/idtoken"
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
func RemoveDupHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		for name, values := range c.Request.Header {
			if len(values) > 1 {
				c.Request.Header[name] = values[:1]
			}
		}
	}
}
func CorsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return cors.New(cors.Config{
		//AllowAllOrigins: true,
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		//AllowOrigins: allowedOrigins,
		//AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		//MaxAge:           12 * time.Hour, // How long the results of a preflight request can be cached
	})
}

func OtelGinMiddleware() gin.HandlerFunc {
	return otelgin.Middleware(os.Getenv("PROD_DOMAIN"))
}

func GcpPubSubAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if os.Getenv("ENVIRONMENT") == "dev" {
			c.Next()
			return
		}
		headerValue := c.GetHeader("Authorization")
		if headerValue == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			c.Abort()
			return
		}
		token := strings.Split(headerValue, "Bearer ")[1]
		_, err := idtoken.Validate(c.Request.Context(), token, "https://"+os.Getenv("PROD_DOMAIN")+"/blog/events/posts-updated")
		if err != nil {
			slog.ErrorContext(c.Request.Context(), "Invalid ID token", "Error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid ID token"})
			c.Abort()
			return
		}
		c.Next()
	}
}
