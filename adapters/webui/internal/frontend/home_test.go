package frontend_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luno/workflow/adapters/webui/internal/frontend"
)

func TestHomeHandlerFuncPaths(t *testing.T) {
	paths := frontend.Paths{
		List:       "/api/list",
		Update:     "/api/update",
		ObjectData: "/api/object",
		Graph:      "/api/graph",
		Steps:      "/api/steps",
	}

	rec := httptest.NewRecorder()
	frontend.HomeHandlerFunc(paths)(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, `list:       "/api/list",`)
	require.Contains(t, body, `objectData: "/api/object",`)
	require.Contains(t, body, `graph:      "/api/graph",`)
	require.Contains(t, body, `steps:      "/api/steps",`)
	require.Contains(t, body, `timeline:   ""`)
}
