package password_test

import (
	"testing"

	"github.com/golang-api-server/pkg/password"
	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	hash, err := password.Hash("mypassword")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "mypassword", hash)
}

func TestCompare_Valid(t *testing.T) {
	hash, err := password.Hash("mypassword")
	assert.NoError(t, err)

	ok := password.Compare(hash, "mypassword")
	assert.True(t, ok)
}

func TestCompare_Invalid(t *testing.T) {
	hash, err := password.Hash("mypassword")
	assert.NoError(t, err)

	ok := password.Compare(hash, "wrongpassword")
	assert.False(t, ok)
}
