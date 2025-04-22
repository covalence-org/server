package tests

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
)

func TestSignInWithEmailPassword(t *testing.T) {
	// load your ../.env, if present
	if err := godotenv.Load("../.env"); err != nil {
		t.Logf("warning: could not load .env: %v", err)
	}

	// pull in the same env vars your main uses
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("SUPABASE_URL or SUPABASE_KEY not set; skipping integration test")
	}

	// initialize client
	client, err := supabase.NewClient(supabaseURL, supabaseKey, &supabase.ClientOptions{})
	if err != nil {
		t.Fatalf("failed to initialize Supabase client: %v", err)
	}

	// perform the same sign‑in call
	resp, err := client.Auth.SignInWithEmailPassword(
		"alialhamadani72@gmail.com",
		"password123",
	)
	if err != nil {
		t.Fatalf("error signing in user: %v", err)
	}

	// basic assertion: we got back a non‑empty access token
	if resp.AccessToken == "" {
		t.Errorf("expected a non-empty AccessToken, got %#v", resp)
	}

	// optional: log it for human inspection
	log.Printf("logged in, got user ID: %s", resp.User.ID)
}
