package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luno/workflow/adapters/webui/internal/api"
)

// postRun posts a RunRequest for the run ID to the handler and returns the response.
func postRun(t *testing.T, handler http.HandlerFunc, runID string) *httptest.ResponseRecorder {
	body, err := json.Marshal(api.RunRequest{RunID: runID})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body)))
	return rec
}
