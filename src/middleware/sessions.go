package middleware

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	sessionRedis "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

const DefaultSessionName = "covalence_session"

// SessionHandler creates session middleware using Redis.
// Expects SESSION_SECRET_KEY env var.
func SessionHandler() gin.HandlerFunc {
	secret := os.Getenv("SESSION_SECRET_KEY")
	if secret == "" {
		if gin.Mode() == gin.DebugMode {
			log.Println("warning: SESSION_SECRET_KEY not set. Using insecure default for debug mode.")
			secret = "dev-secret-needs-changing" // Insecure default ONLY for debug
		} else {
			log.Fatal("fatal: SESSION_SECRET_KEY environment variable not set!")
		}
	}

	// Use the global Redis client initialized earlier
	if RDB == nil {
		log.Fatal("fatal: redis client (RDB) not initialized before creating session handler.")
	}
	// Note: sessionRedis expects the go-redis client interface.
	// Ensure your RDB type implements the necessary methods. go-redis/v9 should work.
	store, err := sessionRedis.NewStore(10, "tcp", RDB.Options().Addr, "", RDB.Options().Password, []byte(secret))
	if err != nil {
		log.Fatalf("fatal: failed to create redis session store: %v", err)
	}

	// Configure session options for production
	store.Options(sessions.Options{
		Path:     "/",
		HttpOnly: true,                        // Must be true for security
		Secure:   gin.Mode() != gin.DebugMode, // True in production (requires HTTPS)
		MaxAge:   86400 * 7,                   // 7 days
		SameSite: http.SameSiteLaxMode,        // Good balance; Strict is more secure but can break some cross-origin navigations.
	})

	return sessions.Sessions(DefaultSessionName, store)
}
