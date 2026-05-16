package awslambda

import (
	"context"
	"fmt"
	"os"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

const accessJWTHeaderKey = "cf-access-jwt-assertion"

type accessClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type accessVerifier struct {
	jwks     keyfunc.Keyfunc
	issuer   string
	audience string
}

var verifier = mustInitVerifier()

func mustInitVerifier() *accessVerifier {
	teamDomain := os.Getenv("CF_ACCESS_TEAM_DOMAIN")
	aud := os.Getenv("CF_ACCESS_AUD")
	if teamDomain == "" || aud == "" {
		panic("CF_ACCESS_TEAM_DOMAIN and CF_ACCESS_AUD must be set")
	}

	issuer := fmt.Sprintf("https://%s.cloudflareaccess.com", teamDomain)
	jwksURL := issuer + "/cdn-cgi/access/certs"

	jwks, err := keyfunc.NewDefaultCtx(context.Background(), []string{jwksURL})
	if err != nil {
		panic(fmt.Sprintf("failed to fetch JWKS from %s: %v", jwksURL, err))
	}

	return &accessVerifier{
		jwks:     jwks,
		issuer:   issuer,
		audience: aud,
	}
}

func (v *accessVerifier) verify(token string) (string, error) {
	claims := &accessClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, v.jwks.Keyfunc,
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithValidMethods([]string{"RS256"}),
	)
	if err != nil {
		return "", fmt.Errorf("invalid Access JWT: %w", err)
	}
	if !parsed.Valid {
		return "", fmt.Errorf("Access JWT marked invalid")
	}
	if claims.Email == "" {
		return "", fmt.Errorf("Access JWT missing email claim")
	}
	return claims.Email, nil
}

type emailContextKey struct{}

func contextWithEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, emailContextKey{}, email)
}

func emailFromContext(ctx context.Context) string {
	email, _ := ctx.Value(emailContextKey{}).(string)
	return email
}
