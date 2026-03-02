package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContains(t *testing.T) {
	assert.True(t, Contains([]string{"a", "b"}, "a"))
	assert.False(t, Contains([]string{"a", "b"}, "c"))
	assert.False(t, Contains(nil, "a"))
}

func TestValidOrderStatusesCoverAPI(t *testing.T) {
	for _, s := range []string{"pending", "open", "partial_fill", "filled", "cancelled", "rejected", "all"} {
		assert.True(t, Contains(ValidOrderStatuses, s), "missing status %q", s)
	}
}

func TestReturnErrorJSON(t *testing.T) {
	w := httptest.NewRecorder()
	ReturnErrorJSON(w, "something broke", http.StatusTeapot)

	assert.Equal(t, http.StatusTeapot, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "something broke", body["message"])
}

func TestExtractBearerToken(t *testing.T) {
	newReq := func(header string) *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		if header != "" {
			r.Header.Set("Authorization", header)
		}
		return r
	}

	cases := []struct {
		name    string
		header  string
		want    string
		wantErr string
	}{
		{"valid", "Bearer tok-123", "tok-123", ""},
		{"case-insensitive scheme", "bearer tok-123", "tok-123", ""},
		{"missing header", "", "", "missing authorization header"},
		{"wrong scheme", "Basic dXNlcjpwYXNz", "", "invalid authorization format"},
		{"no token part", "Bearer", "", "invalid authorization format"},
		{"empty token", "Bearer   ", "", "empty token"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := ExtractBearerToken(newReq(tc.header))
			if tc.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.want, token)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}
