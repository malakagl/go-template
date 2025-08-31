package util

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/malakagl/kart-challenge/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// AbsoluteFilePath constructs an absolute file path based on relative path.
// The function returns the absolute path to the specified file.
func AbsoluteFilePath(file, relativePath string) string {
	_, thisFile, _, _ := runtime.Caller(0) // 0 = this function
	baseDir := filepath.Join(filepath.Dir(thisFile), relativePath)
	return filepath.Join(baseDir, file)
}

func StringToUint(s string) (uint, error) {
	t, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}

	return uint(t), nil
}

func MapErrorToHTTP(err error) (int, string) {
	switch {
	case errors.Is(err, errors.ErrInvalidCouponCode):
		return http.StatusUnprocessableEntity, "Invalid coupon code"
	default:
		return http.StatusInternalServerError, "Failed to create order"
	}
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

// GenerateAPIKey returns (clientID, secretHash, fullKey)
func GenerateAPIKey() (string, string, string, error) {
	// Public client ID
	clientIDBytes, _ := generateRandomBytes(8)
	clientID := base64.RawURLEncoding.EncodeToString(clientIDBytes)

	// Secret value
	secretBytes, _ := generateRandomBytes(32)
	secret := base64.RawURLEncoding.EncodeToString(secretBytes)

	// Hash the secret
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", "", "", err
	}

	// Full key (what you return to client)
	fullKey := fmt.Sprintf("%s.%s", clientID, secret)

	return clientID, string(hash), fullKey, nil
}
