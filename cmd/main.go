package main

import (
	"context"
	"log"
	"net/http"
	"os"

	traefik_entra_sso_plugin "github.com/KenMusembi/traefik-entra-sso-plugin/plugin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	tenantID := os.Getenv("AZURE_TENANT_ID")
	clientID := os.Getenv("AZURE_CLIENT_ID")
	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")

	if tenantID == "" || clientID == "" || clientSecret == "" {
		log.Fatalf("TenantID, ClientID, and ClientSecret cannot be empty")
	}

	ctx := context.Background()
	config := &traefik_entra_sso_plugin.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	handler, err := traefik_entra_sso_plugin.New(ctx, http.DefaultServeMux, config, "entra-sso")
	if err != nil {
		log.Fatalf("Error creating traefik_entra_sso_plugin handler: %v", err)
	}

	http.Handle("/", handler)
	log.Println("Starting server on :8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
