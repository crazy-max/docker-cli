package hub

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/assert"
)

func TestDoRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Header.Get("Accept"), "application/json")
		assert.Equal(t, r.Header.Get("User-Agent"), testUserAgent())
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	assert.NilError(t, err)

	c := NewClient(testUserAgent())
	_, err = c.doRequest(req)
	assert.NilError(t, err)
}
