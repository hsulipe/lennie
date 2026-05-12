package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/hsulipe/lennie/login/internal/adapters/httphandler"
	pgadapter "github.com/hsulipe/lennie/login/internal/adapters/postgres"
	"github.com/hsulipe/lennie/login/internal/services"
	"golang.org/x/oauth2"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func connectDatabase() *gorm.DB {
	dsn := "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable"
	db, err := gorm.Open(
		gormpostgres.Open(fmt.Sprintf(
			dsn,
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
		)),
		&gorm.Config{},
	)
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

// initOIDC sets up the OAuth2 config and ID token verifier.
//
// OIDC_ISSUER is the public-facing issuer URL (used by the browser and for
// token claim validation). OIDC_PROVIDER_URL overrides where the server fetches
// the discovery document, JWKS, and token endpoint — set this to the Docker
// service name (e.g. http://dex:5556/dex) so the container can reach Dex
// directly without going through the host. If unset, OIDC_ISSUER is used.
func initOIDC(ctx context.Context) (*oidc.IDTokenVerifier, *oauth2.Config) {
	issuer := os.Getenv("OIDC_ISSUER")
	providerURL := os.Getenv("OIDC_PROVIDER_URL")
	if providerURL == "" {
		providerURL = issuer
	}

	// Fetch the discovery document from the internal URL while accepting that
	// its "issuer" field will contain the public URL.
	fetchCtx := oidc.InsecureIssuerURLContext(ctx, issuer)
	provider, err := oidc.NewProvider(fetchCtx, providerURL)
	if err != nil {
		log.Fatalf("failed to initialize OIDC provider from %s: %v", providerURL, err)
	}

	// The discovery doc's endpoints use the public issuer URL. When a separate
	// providerURL is set, swap the host in the token endpoint so the container
	// calls Dex directly rather than routing through the host machine.
	endpoint := provider.Endpoint()
	endpoint.TokenURL = strings.Replace(endpoint.TokenURL, issuer, providerURL, 1)

	// Build the verifier with a JWKS URL that points to the internal address,
	// while still checking the "iss" claim against the public issuer URL.
	jwksURL := strings.TrimRight(providerURL, "/") + "/keys"
	keySet := oidc.NewRemoteKeySet(ctx, jwksURL)
	verifier := oidc.NewVerifier(issuer, keySet, &oidc.Config{
		ClientID: os.Getenv("OIDC_CLIENT_ID"),
	})

	cfg := &oauth2.Config{
		ClientID:     os.Getenv("OIDC_CLIENT_ID"),
		ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OIDC_REDIRECT_URL"),
		Endpoint:     endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	return verifier, cfg
}

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	db := connectDatabase()
	repo := pgadapter.NewUserRepository(db)

	oidcVerifier, oidcConfig := initOIDC(context.Background())

	signInSvc := services.NewSignInService(repo)
	signUpSvc := services.NewSignUpService(repo)
	oidcSvc := services.NewOIDCService(repo)

	handler := httphandler.NewHandler(signInSvc, signUpSvc, oidcSvc, oidcConfig, oidcVerifier, db != nil)
	router := httphandler.NewRouter(handler)

	fmt.Println("Server is running on http://localhost:" + port)
	http.ListenAndServe(":"+port, router)
}
