package jwt_test

import (
	"testing"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/golang-api-server/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key"

func TestGenerateTokenPair(t *testing.T) {
	pair, err := jwt.GenerateTokenPair(
		testSecret,
		1,
		"test@example.com",
		"user",
		15*time.Minute,
		7*24*time.Hour,
	)

	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.True(t, pair.ExpiresAt.After(time.Now()))
}

func TestValidateToken_Valid(t *testing.T) {
	pair, err := jwt.GenerateTokenPair(
		testSecret,
		42,
		"user@test.com",
		"admin",
		15*time.Minute,
		7*24*time.Hour,
	)
	require.NoError(t, err)

	claims, err := jwt.ValidateToken(testSecret, pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, int64(42), claims.UserID)
	assert.Equal(t, "user@test.com", claims.Email)
	assert.Equal(t, "admin", claims.Role)
}

func TestValidateToken_InvalidSecret(t *testing.T) {
	pair, err := jwt.GenerateTokenPair(
		testSecret,
		1,
		"user@test.com",
		"user",
		15*time.Minute,
		7*24*time.Hour,
	)
	require.NoError(t, err)

	_, err = jwt.ValidateToken("wrong-secret", pair.AccessToken)
	assert.ErrorIs(t, err, jwt.ErrTokenInvalid)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	_, err := jwt.ValidateToken(testSecret, "not-a-token")
	assert.ErrorIs(t, err, jwt.ErrTokenInvalid)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	claims := &jwt.Claims{
		UserID: 1,
		Email:  "test@example.com",
		Role:   "user",
	}
	claims.ExpiresAt = gojwt.NewNumericDate(time.Now().Add(-time.Hour))
	claims.IssuedAt = gojwt.NewNumericDate(time.Now().Add(-2 * time.Hour))
	claims.Issuer = "golang-api-server"
	claims.Subject = "1"

	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSecret))
	require.NoError(t, err)

	_, err = jwt.ValidateToken(testSecret, tokenString)
	assert.ErrorIs(t, err, jwt.ErrTokenExpired)
}
