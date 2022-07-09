package medium

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAction_Status(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()

	action := NewAction(rw, r, nil)
	action.ResponseWriter().WriteHeader(http.StatusNotFound)

	require.Equal(t, 404, action.Status())
}
