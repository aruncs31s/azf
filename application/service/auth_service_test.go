package service

import (
	"os"
	"testing"
	"time"
)

func TestGetJWTSecret_NotSet(t *testing.T) {
	os.Unsetenv("JWT_SECRET")
	_, err := GetJWTSecret()
	if err != ErrJWTSecretNotSet {
		t.Errorf("Expected ErrJWTSecretNotSet, got %v", err)
	}
}

func TestGetJWTSecret_TooShort(t *testing.T) {
	os.Setenv("JWT_SECRET", "short")
	defer os.Unsetenv("JWT_SECRET")

	_, err := GetJWTSecret()
	if err != ErrJWTSecretTooShort {
		t.Errorf("Expected ErrJWTSecretTooShort, got %v", err)
	}
}

func TestGetJWTSecret_Valid(t *testing.T) {
	validSecret := "this-is-a-very-long-secret-key-that-is-at-least-32-characters"
	os.Setenv("JWT_SECRET", validSecret)
	defer os.Unsetenv("JWT_SECRET")

	secret, err := GetJWTSecret()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if secret != validSecret {
		t.Errorf("Expected secret to be '%s', got '%s'", validSecret, secret)
	}
}

func TestGenerateToken(t *testing.T) {
	validSecret := "this-is-a-very-long-secret-key-that-is-at-least-32-characters"
	os.Setenv("JWT_SECRET", validSecret)
	defer os.Unsetenv("JWT_SECRET")

	claims := map[string]any{
		"user_id": "123",
		"role":    "admin",
	}

	token, err := GenerateToken(claims)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if token == "" {
		t.Error("Expected token to be non-empty")
	}
}

func TestGenerateToken_NoSecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET")

	claims := map[string]any{
		"user_id": "123",
	}

	_, err := GenerateToken(claims)
	if err != ErrJWTSecretNotSet {
		t.Errorf("Expected ErrJWTSecretNotSet, got %v", err)
	}
}

func TestGenerateAndValidateToken(t *testing.T) {
	validSecret := "this-is-a-very-long-secret-key-that-is-at-least-32-characters"
	os.Setenv("JWT_SECRET", validSecret)
	defer os.Unsetenv("JWT_SECRET")

	claims := map[string]any{
		"user_id": "123",
		"role":    "admin",
	}

	token, err := GenerateToken(claims)
	if err != nil {
		t.Errorf("Expected no error generating token, got %v", err)
	}

	validatedClaims, err := ValidateJWT(token)
	if err != nil {
		t.Errorf("Expected no error validating token, got %v", err)
	}

	if validatedClaims["user_id"] != "123" {
		t.Errorf("Expected user_id to be '123', got '%v'", validatedClaims["user_id"])
	}
	if validatedClaims["role"] != "admin" {
		t.Errorf("Expected role to be 'admin', got '%v'", validatedClaims["role"])
	}
}

func TestGenerateAccessToken(t *testing.T) {
	validSecret := "this-is-a-very-long-secret-key-that-is-at-least-32-characters"
	os.Setenv("JWT_SECRET", validSecret)
	defer os.Unsetenv("JWT_SECRET")

	claims := map[string]any{
		"user_id": "123",
	}

	token, err := GenerateAccessToken(claims)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if token == "" {
		t.Error("Expected token to be non-empty")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	validSecret := "this-is-a-very-long-secret-key-that-is-at-least-32-characters"
	os.Setenv("JWT_SECRET", validSecret)
	defer os.Unsetenv("JWT_SECRET")

	claims := map[string]any{
		"user_id": "123",
	}

	token, err := GenerateRefreshToken(claims)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if token == "" {
		t.Error("Expected token to be non-empty")
	}
}

func TestGenerateJWT_CustomExpiry(t *testing.T) {
	secret := "test-secret-key-for-testing-purposes-only"
	claims := MapToClaims(map[string]any{
		"user_id": "123",
	})

	token, err := GenerateJWT(secret, claims, 1*time.Hour)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if token == "" {
		t.Error("Expected token to be non-empty")
	}
}

func TestMapToClaims(t *testing.T) {
	data := map[string]interface{}{
		"user_id": "123",
		"role":    "admin",
		"active":  true,
	}

	claims := MapToClaims(data)

	if claims["user_id"] != "123" {
		t.Errorf("Expected user_id to be '123', got '%v'", claims["user_id"])
	}
	if claims["role"] != "admin" {
		t.Errorf("Expected role to be 'admin', got '%v'", claims["role"])
	}
	if claims["active"] != true {
		t.Errorf("Expected active to be true, got '%v'", claims["active"])
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	value := GetEnv("TEST_VAR", "fallback")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}

	value = GetEnv("NON_EXISTENT_VAR", "fallback")
	if value != "fallback" {
		t.Errorf("Expected 'fallback', got '%s'", value)
	}
}
