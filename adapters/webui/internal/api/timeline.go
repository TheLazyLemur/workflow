package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/luno/workflow"
)

// ListRunEvents returns the events that the workflow published for a run.
type ListRunEvents func(ctx context.Context, runID string) ([]workflow.Event, error)

type TimelineResponse struct {
	Items []TimelineItem `json:"items"`
}

// TimelineItem is a single saved version of a run along with the events published for it.
type TimelineItem struct {
	Version   uint            `json:"version"`
	Status    string          `json:"status"`
	RunState  string          `json:"run_state"`
	UpdatedAt time.Time       `json:"updated_at"`
	Events    []TimelineEvent `json:"events"`
}

type TimelineEvent struct {
	ID        int64     `json:"id"`
	Topic     string    `json:"topic"`
	CreatedAt time.Time `json:"created_at"`
}

// Timeline serves every saved version of a run, oldest first. When listEvents is not nil each version lists the
// events published for it, matched on the record version header of the event.
func Timeline(listHistory ListHistoryFn, listEvents ListRunEvents) http.HandlerFunc {
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

		resp := TimelineResponse{Items: []TimelineItem{}}
		byVersion := make(map[uint]int)
		for _, version := range history {
			byVersion[version.Meta.Version] = len(resp.Items)
			resp.Items = append(resp.Items, TimelineItem{
				Version:   version.Meta.Version,
				Status:    version.Meta.StatusDescription,
				RunState:  version.RunState.String(),
				UpdatedAt: version.UpdatedAt,
				Events:    []TimelineEvent{},
			})
		}

		if listEvents != nil {
			events, err := listEvents(r.Context(), req.RunID)
			if err != nil {
				http.Error(w, "failed to list events of run", http.StatusInternalServerError)
				return
			}

			for _, event := range events {
				// Every event published by the workflow carries the version of the record it was published for.
				header := event.Headers[workflow.HeaderRecordVersion]
				version, err := strconv.ParseUint(header, 10, 0)
				i, ok := byVersion[uint(version)]
				if err != nil || !ok {
					msg := fmt.Sprintf("event %d has no version in the history of the run (record version %q)", event.ID, header)
					http.Error(w, msg, http.StatusInternalServerError)
					return
				}

				resp.Items[i].Events = append(resp.Items[i].Events, TimelineEvent{
					ID:        event.ID,
					Topic:     event.Headers[workflow.HeaderTopic],
					CreatedAt: event.CreatedAt,
				})
			}
		}

		writeJSON(w, resp)
	}
}
