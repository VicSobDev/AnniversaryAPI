package main

import (
	"log"
	"os"

	"github.com/VicSobDev/anniversaryAPI/internal/server"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {

	log.Println("Starting server...")

	jwtKey := os.Getenv("JWT_KEY")
	if jwtKey == "" {
		log.Fatal("JWT_KEY environment variable is not set")
	}

	prometheusKey := os.Getenv("PROMETHEUS_KEY")
	if prometheusKey == "" {
		log.Fatal("PROMETHEUS_KEY environment variable is not set")
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY environment variable is not set")
	}

	api := server.NewApi(":8080", []byte(jwtKey), prometheusKey, apiKey)

	if err := api.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
