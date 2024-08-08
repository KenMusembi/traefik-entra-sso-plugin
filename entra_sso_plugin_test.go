package main_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/coreos/go-oidc"
)

// Config holds the configuration for the EntraSSO plugin.
type Config struct {
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
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
		oidcProvider, err = oidc.NewProvider(ctx, "https://login.microsoftonline.com/"+config.ClientID+"/v2.0")
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
	return oidc.NewVerifier("https://login.microsoftonline.com/382f5ca8-4f68-4dea-b9a3-358d26417a8e/v2.0", mockKeySet, config)
}

// TestEntraSSO tests the EntraSSO plugin's functionality.
func TestEntraSSO(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("AZURE_TENANT_ID", "test-tenant-id")
	os.Setenv("AZURE_CLIENT_ID", "test-client-id")
	os.Setenv("AZURE_CLIENT_SECRET", "test-client-secret")

	// Create a mock OIDC provider
	mockProvider := &MockProvider{}

	// Create the plugin configuration
	config := &Config{
		ClientID:     os.Getenv("AZURE_CLIENT_ID"),
		ClientSecret: os.Getenv("AZURE_CLIENT_SECRET"),
	}

	// Create the EntraSSO plugin instance
	handler, err := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), config, "entra-sso", mockProvider)

	if err != nil {
		t.Fatalf("Failed to create EntraSSO plugin: %v", err)
	}

	// Create a request with a valid Bearer token
	req := httptest.NewRequest(http.MethodGet, "http://hloveafrica.org", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	// Record the response
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check that the status code is 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusOK)
	}

	// Create a request with an invalid Bearer token
	req = httptest.NewRequest(http.MethodGet, "http://hloveafrica.org", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	// Record the response
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check that the status code is 401 Unauthorized
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusUnauthorized)
	}

	// Create a request without an Authorization header
	req = httptest.NewRequest(http.MethodGet, "http://hloveafrica.org", nil)

	// Record the response
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check that the status code is 401 Unauthorized
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusUnauthorized)
	}
}
