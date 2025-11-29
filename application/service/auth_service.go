package service

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Token expiry constants
const (
	// DefaultAccessTokenExpiry is the default access token expiry
	DefaultAccessTokenExpiry = 15 * time.Minute
	// DefaultRefreshTokenExpiry is the default refresh token expiry
	DefaultRefreshTokenExpiry = 7 * 24 * time.Hour
	// MinJWTSecretLength is the minimum required length for JWT secret
	MinJWTSecretLength = 32
)

// JWT-related errors
var (
	ErrJWTSecretNotSet   = errors.New("JWT_SECRET environment variable is not set")
	ErrJWTSecretTooShort = errors.New("JWT_SECRET must be at least 32 characters")
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token has expired")
)

// GetJWTSecret retrieves and validates the JWT secret from environment
func GetJWTSecret() (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", ErrJWTSecretNotSet
	}
	if len(secret) < MinJWTSecretLength {
		return "", ErrJWTSecretTooShort
	}
	return secret, nil
}

// GenerateToken generates a JWT token with the given claims
// Uses secure defaults for token expiry
func GenerateToken(claims map[string]any) (string, error) {
	secret, err := GetJWTSecret()
	if err != nil {
		return "", err
	}
	jwtClaims := MapToClaims(claims)
	return GenerateJWT(secret, jwtClaims, DefaultAccessTokenExpiry)
}

// GenerateAccessToken generates a short-lived access token
func GenerateAccessToken(claims map[string]any) (string, error) {
	return GenerateTokenWithExpiry(claims, DefaultAccessTokenExpiry)
}

// GenerateRefreshToken generates a long-lived refresh token
func GenerateRefreshToken(claims map[string]any) (string, error) {
	return GenerateTokenWithExpiry(claims, DefaultRefreshTokenExpiry)
}

// GenerateTokenWithExpiry generates a token with custom expiry
func GenerateTokenWithExpiry(claims map[string]any, expiry time.Duration) (string, error) {
	secret, err := GetJWTSecret()
	if err != nil {
		return "", err
	}
	jwtClaims := MapToClaims(claims)
	return GenerateJWT(secret, jwtClaims, expiry)
}

func GetEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

// GenerateJWT generates a JWT token with the given claims and secret
func GenerateJWT(secret string, claims jwt.MapClaims, expiry time.Duration) (string, error) {
	if claims == nil {
		claims = jwt.MapClaims{}
	}
	now := time.Now()
	claims["exp"] = now.Add(expiry).Unix()
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT validates the token string using the secret and returns the claims
func ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	secret, err := GetJWTSecret()
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrInvalidToken
}

// MapToClaims converts a map to jwt.MapClaims
func MapToClaims(data map[string]interface{}) jwt.MapClaims {
	claims := jwt.MapClaims{}
	for k, v := range data {
		claims[k] = v
	}
	return claims
}
