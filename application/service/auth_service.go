package service

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(claims map[string]any) (string, error) {
	jwtClaims := MapToClaims(claims)
	secret := GetEnv("JWT_SECRET", "")
	return GenerateJWT(secret, jwtClaims, 1000*time.Hour)
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
	claims["exp"] = time.Now().Add(expiry).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT validates the token string using the secret and returns the claims
// func ValidateJWT(tokenString, secret string) (jwt.MapClaims, error) {
// 	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 		if token.Method != jwt.SigningMethodHS256 {
// 			return nil, errors.New("unexpected signing method")
// 		}
// 		return []byte(secret), nil
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
// 		return claims, nil
// 	}
// 	return nil, errors.New("invalid token")
// }

// MapToClaims converts a map to jwt.MapClaims
func MapToClaims(data map[string]interface{}) jwt.MapClaims {
	claims := jwt.MapClaims{}
	for k, v := range data {
		claims[k] = v
	}
	return claims
}
