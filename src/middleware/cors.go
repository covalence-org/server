package middleware

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CorsHandler creates CORS middleware configured via environment variable.
func CorsHandler() gin.HandlerFunc {
	originsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	var allowedOrigins []string
	if originsEnv == "" {
		log.Println("warning: CORS_ALLOWED_ORIGINS not set. allowing all origins (INSECURE for production).")
		// Fallback for local dev or if explicitly desired, but insecure
		return cors.Default() // Less secure default, allows all origins/methods
	} else {
		allowedOrigins = strings.Split(originsEnv, ",")
		// Trim whitespace just in case
		for i := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
		}
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"}, // Add any other custom headers needed
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true, // Crucial for sessions/cookies
		MaxAge:           12 * time.Hour,
	})
}
