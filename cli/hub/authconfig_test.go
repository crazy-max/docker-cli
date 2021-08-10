package hub

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestAuthConfig(t *testing.T) {
	_, err := AuthConfig()
	assert.NilError(t, err)
}
