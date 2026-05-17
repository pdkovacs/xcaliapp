package awslambda

import (
	"context"
	"fmt"
	"os"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

const accessJWTHeaderKey = "cf-access-jwt-assertion"

// devAuthEmail is the identity attributed to every request when the Cloudflare
// Access gate is bypassed via XCALI_DEV_AUTH=skip (local/LocalStack testing
// only). It is never used when the flag is unset.
const devAuthEmail = "dev@local"

// devAuthSkip reports whether Cloudflare Access verification should be skipped.
// This is strictly opt-in via XCALI_DEV_AUTH=skip and must never be set in
// production.
func devAuthSkip() bool {
	return os.Getenv("XCALI_DEV_AUTH") == "skip"
}

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
	if devAuthSkip() {
		// Skip the JWKS fetch and the required-env-var check entirely; the
		// verifier is never consulted while XCALI_DEV_AUTH=skip is set.
		return nil
	}

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
