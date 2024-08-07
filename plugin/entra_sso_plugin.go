package traefik_entra_sso_plugin

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/coreos/go-oidc"
)

// Config holds the plugin configuration.
type Config struct {
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// EntraSSO is a Traefik plugin for Microsoft Entra SSO.
type EntraSSO struct {
	next   http.Handler
	name   string
	config *Config
}

// New creates a new EntraSSO plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	tenantID := os.Getenv("AZURE_TENANT_ID")
	clientID := os.Getenv("AZURE_CLIENT_ID")
	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")

	if tenantID == "" || clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("TenantID, ClientID, and ClientSecret cannot be empty")
	}

	return &EntraSSO{
		next:   next,
		name:   name,
		config: config,
	}, nil
}

func (e *EntraSSO) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := context.Background()

	tenantID := os.Getenv("AZURE_TENANT_ID")
	clientID := os.Getenv("AZURE_CLIENT_ID")

	provider, err := oidc.NewProvider(ctx, fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", tenantID))
	if err != nil {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	rawIDToken, err := getRawIDToken(req)
	if err != nil {
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	e.next.ServeHTTP(rw, req)
}

func getRawIDToken(req *http.Request) (string, error) {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is missing")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("invalid authorization header")
	}

	return strings.TrimPrefix(authHeader, "Bearer "), nil
}
