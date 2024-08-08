package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-oidc"
)

// Config holds the configuration for the EntraSSO plugin.
type Config struct {
	ClientID     string
	ClientSecret string
	TenantID     string
}

// Provider is an interface that both oidc.Provider and MockProvider implement.
type Provider interface {
	Verifier(config *oidc.Config) *oidc.IDTokenVerifier
}

// EntraSSO is the main struct for the plugin.
type EntraSSO struct {
	next     http.Handler
	provider Provider
	verifier *oidc.IDTokenVerifier
	config   *Config
}

// New creates a new instance of EntraSSO plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string, provider Provider) (http.Handler, error) {
	if provider == nil {
		var err error
		var oidcProvider *oidc.Provider
		oidcProvider, err = oidc.NewProvider(ctx, "https://login.microsoftonline.com/"+config.TenantID+"/v2.0")
		if err != nil {
			return nil, err
		}
		provider = oidcProvider
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: config.ClientID})
	return &EntraSSO{
		next:     next,
		provider: provider,
		verifier: verifier,
		config:   config,
	}, nil
}

// ServeHTTP is the main handler function for the EntraSSO plugin.
func (a *EntraSSO) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Example authentication logic (to be replaced with real logic).
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) < 8 {
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}

	token := authHeader[7:]
	_, err := a.verifier.Verify(req.Context(), token)
	if err != nil {
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// This is where you would process the token and authenticate the user.
	// For now, just proceed to the next handler.
	a.next.ServeHTTP(rw, req)
}

// MockKeySet implements the oidc.KeySet interface for testing purposes.
type MockKeySet struct{}

// VerifySignature mocks the signature verification process.
func (m *MockKeySet) VerifySignature(ctx context.Context, jwt string) ([]byte, error) {
	if jwt == "valid-token" {
		return []byte(jwt), nil
	}
	return nil, errors.New("invalid token")
}

// MockProvider is a mock implementation of the Provider interface.
type MockProvider struct{}

func (m *MockProvider) Verifier(config *oidc.Config) *oidc.IDTokenVerifier {
	mockKeySet := &MockKeySet{}
	return oidc.NewVerifier("https://login.microsoftonline.com/test-tenant-id/v2.0", mockKeySet, config)
}

func main() {
	// Load configuration
	config := &Config{
		ClientID:     os.Getenv("AZURE_CLIENT_ID"),
		ClientSecret: os.Getenv("AZURE_CLIENT_SECRET"),
		TenantID:     os.Getenv("AZURE_TENANT_ID"), // Ensure this is set correctly
	}

	// Create a new EntraSSO plugin instance
	var provider Provider
	var err error
	if os.Getenv("MOCK_PROVIDER") == "true" {
		provider = &MockProvider{}
	} else {
		provider, err = oidc.NewProvider(context.Background(), "https://login.microsoftonline.com/"+config.TenantID+"/v2.0")
		if err != nil {
			log.Fatalf("Failed to create OIDC provider: %v", err)
		}
	}

	handler, err := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "entra-sso", provider)

	if err != nil {
		log.Fatalf("Failed to create EntraSSO plugin: %v", err)
	}

	http.Handle("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
