package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
)

func main() {
	// Load .env (contains SUPABASE_URL and SUPABASE_KEY)
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("Warning: could not load .env file: %v", err)
	}

	// Use your service_role key here for admin operations
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	// Initialize Supabase client :contentReference[oaicite:0]{index=0}
	client, err := supabase.NewClient(supabaseURL, supabaseKey, &supabase.ClientOptions{})
	if err != nil {
		log.Fatalf("Failed to initialize Supabase client: %v", err)
	}

	// req := types.SignupRequest{
	// 	Email:    "alialhamadani72@gmail.com",
	// 	Password: "password123",
	// }

	// // Perform the request with a context
	// resp, err := client.Auth.Signup(req)
	// if err != nil {
	// 	log.Fatalf("Error creating user: %v", err)
	// }

	// // Output the result
	// fmt.Printf("New user created: %+v\n", resp)

	// Perform the request with a context
	resp, err := client.Auth.SignInWithEmailPassword(
		"alialhamadani72@gmail.com",
		"password123",
	)
	if err != nil {
		log.Fatalf("Error signing in user: %v", err)
	}

	// Output the result
	fmt.Printf("User logged in: %+v\n", resp)

}
