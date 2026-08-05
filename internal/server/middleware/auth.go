package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/sahil/leasewebassignment/internal/api"
)

var (
	ErrInvalidToken = errors.New("invalid authorization token")
	ErrTokenExpired = errors.New("authorization token expired")
)

type AuthMiddleware struct {
	jwtSigningKey string
}

func NewAuthMiddleware(jwtSigningKey string) *AuthMiddleware {
	return &AuthMiddleware{jwtSigningKey: jwtSigningKey}
}

func (a *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearerToken(r.Header.Get("Authorization"))
		if token == "" || a.jwtSigningKey == "" {
			api.WriteError(w, api.Unauthorized("authorization required", "authorization header missing or invalid"))
			return
		}
		if err := a.validateJWT(token); err != nil {
			api.WriteError(w, api.Unauthorized("invalid token", err.Error()))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *AuthMiddleware) validateJWT(token string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ErrInvalidToken
	}
	signingInput := parts[0] + "." + parts[1]
	expected := hmac.New(sha256.New, []byte(a.jwtSigningKey))
	_, _ = expected.Write([]byte(signingInput))
	decodedSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return ErrInvalidToken
	}
	if !hmac.Equal(decodedSig, expected.Sum(nil)) {
		return ErrInvalidToken
	}

	return validateJWTClaims(parts[1])
}

func validateJWTClaims(payloadSegment string) error {
	decoded, err := base64.RawURLEncoding.DecodeString(payloadSegment)
	if err != nil {
		return ErrInvalidToken
	}
	var claims struct {
		Exp int64 `json:"exp,omitempty"`
	}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return ErrInvalidToken
	}
	if claims.Exp != 0 && time.Now().Unix() > claims.Exp {
		return ErrTokenExpired
	}
	return nil
}

func extractBearerToken(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
