package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/luno/workflow"
)

// RunRequest is the request body of the handlers that serve a single run.
type RunRequest struct {
	RunID string `json:"run_id"`
}

type ListHistoryFn func(ctx context.Context, runID string) ([]workflow.Record, error)

func readRunRequest(w http.ResponseWriter, r *http.Request) (RunRequest, bool) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request: cannot read body", http.StatusBadRequest)
		return RunRequest{}, false
	}

	var req RunRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Bad Request: cannot unmarshal body", http.StatusBadRequest)
		return RunRequest{}, false
	}

	return req, true
}

func writeJSON(w http.ResponseWriter, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "failed to json marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(b)
}

// objectJSON returns a record's Object as JSON. Objects are JSON unless a custom delete function replaced them with
// something else, in which case they are returned as a string.
func objectJSON(object []byte) json.RawMessage {
	if len(object) == 0 {
		return json.RawMessage("null")
	}

	if json.Valid(object) {
		return object
	}

	b, _ := json.Marshal(string(object))
	return b
}
