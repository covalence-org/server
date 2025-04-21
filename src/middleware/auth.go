package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/supabase-community/gotrue-go/types"
	"github.com/supabase-community/supabase-go"
)

// AuthMiddleware verifies the bearer token and attaches user info (and a Supabase client) to context
func Auth(serviceClient *supabase.Client, anonKey, supabaseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract Bearer token
		authHeader := c.GetHeader("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization bearer token"})
			return
		}

		// Verify token by fetching user via GoTrue
		authClient := serviceClient.Auth.WithToken(token)
		userResp, err := authClient.GetUser()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Attach user ID
		c.Set("userID", userResp.User.ID)

		// Build a Supabase client scoped to this user
		clientOpts := &supabase.ClientOptions{Headers: make(map[string]string)}
		userClient, _ := supabase.NewClient(supabaseURL, anonKey, clientOpts)
		// Update client with the user's JWT for RLS
		session := types.Session{AccessToken: token}
		userClient.UpdateAuthSession(session)
		c.Set("supabaseClient", userClient)

		c.Next()
	}
}
