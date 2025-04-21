package router

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/supabase-community/gotrue-go/types"
	"github.com/supabase-community/supabase-go"
)

func UserLogin(c *gin.Context, serviceClient *supabase.Client) {

	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := serviceClient.Auth.SignInWithEmailPassword(req.Email, req.Password)
	if err != nil {
		log.Printf("error signing in user: %v", err)
		// It's generally better not to Fatalf in a web handler
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign in user. email or password is incorrect."})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func UserSignup(c *gin.Context, serviceClient *supabase.Client) {
	var req types.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := serviceClient.Auth.Signup(req)
	if err != nil {
		log.Printf("error creating user: %v", err)
		// It's generally better not to Fatalf in a web handler
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Respond with the actual response or relevant data
	c.JSON(http.StatusOK, resp) // Example: respond with the signup response
}

func UserRefreshToken(c *gin.Context, serviceClient *supabase.Client) {

	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := serviceClient.Auth.RefreshToken(req.RefreshToken)
	if err != nil {
		log.Printf("error creating user: %v", err)
		// It's generally better not to Fatalf in a web handler
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusOK, resp) // Example: respond with the signup response
}

func UserMe(c *gin.Context) {
	// Get authenticated supabase client
	client := c.MustGet("supabaseClient").(*supabase.Client)

	resp, err := client.Auth.GetUser()
	if err != nil {
		log.Printf("error creating user: %v", err)
		// It's generally better not to Fatalf in a web handler
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusOK, resp) // Example: respond with the signup response
}
