package api

import (
	"encoding/json"
	"net/http"
	"time"
)

type StepsResponse struct {
	Items []StepItem `json:"items"`
}

// StepItem is a single saved version of a run.
type StepItem struct {
	Version   uint            `json:"version"`
	Status    string          `json:"status"`
	RunState  string          `json:"run_state"`
	UpdatedAt time.Time       `json:"updated_at"`
	Object    json.RawMessage `json:"object"`
}

// Steps serves every saved version of a run, oldest first.
func Steps(listHistory ListHistoryFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := readRunRequest(w, r)
		if !ok {
			return
		}

		history, err := listHistory(r.Context(), req.RunID)
		if err != nil {
			http.Error(w, "failed to list history from store", http.StatusInternalServerError)
			return
		}

		resp := StepsResponse{Items: []StepItem{}}
		for _, version := range history {
			resp.Items = append(resp.Items, StepItem{
				Version:   version.Meta.Version,
				Status:    version.Meta.StatusDescription,
				RunState:  version.RunState.String(),
				UpdatedAt: version.UpdatedAt,
				Object:    objectJSON(version.Object),
			})
		}

		writeJSON(w, resp)
	}
}
