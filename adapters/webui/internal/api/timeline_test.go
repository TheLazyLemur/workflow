package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/luno/workflow"
	"github.com/stretchr/testify/require"

	"github.com/luno/workflow/adapters/webui/internal/api"
)

var timelineNow = time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

func timelineHistory(ctx context.Context, runID string) ([]workflow.Record, error) {
	return []workflow.Record{
		{RunState: workflow.RunStateInitiated, UpdatedAt: timelineNow, Meta: workflow.Meta{Version: 1, StatusDescription: "Start"}},
		{RunState: workflow.RunStateCompleted, UpdatedAt: timelineNow, Meta: workflow.Meta{Version: 2, StatusDescription: "End"}},
	}, nil
}

func timelineEvent(id int64, version string) workflow.Event {
	return workflow.Event{
		ID:        id,
		ForeignID: "1",
		Headers: map[workflow.Header]string{
			workflow.HeaderRecordVersion: version,
			workflow.HeaderTopic:         "test-" + version,
		},
		CreatedAt: timelineNow,
	}
}

func TestTimelineHandler(t *testing.T) {
	listEvents := func(ctx context.Context, runID string) ([]workflow.Event, error) {
		require.Equal(t, "1", runID)
		return []workflow.Event{timelineEvent(7, "1"), timelineEvent(8, "2")}, nil
	}

	resp := postRun(t, api.Timeline(timelineHistory, listEvents), "1")
	require.Equal(t, http.StatusOK, resp.Code)

	var actual api.TimelineResponse
	err := json.Unmarshal(resp.Body.Bytes(), &actual)
	require.NoError(t, err)

	expected := api.TimelineResponse{
		Items: []api.TimelineItem{
			{
				Version:   1,
				Status:    "Start",
				RunState:  "Initiated",
				UpdatedAt: timelineNow,
				Events:    []api.TimelineEvent{{ID: 7, Topic: "test-1", CreatedAt: timelineNow}},
			},
			{
				Version:   2,
				Status:    "End",
				RunState:  "Completed",
				UpdatedAt: timelineNow,
				Events:    []api.TimelineEvent{{ID: 8, Topic: "test-2", CreatedAt: timelineNow}},
			},
		},
	}
	require.Equal(t, expected, actual)
}

func TestTimelineHandlerWithoutEvents(t *testing.T) {
	resp := postRun(t, api.Timeline(timelineHistory, nil), "1")
	require.Equal(t, http.StatusOK, resp.Code)

	var actual api.TimelineResponse
	err := json.Unmarshal(resp.Body.Bytes(), &actual)
	require.NoError(t, err)

	expected := api.TimelineResponse{
		Items: []api.TimelineItem{
			{Version: 1, Status: "Start", RunState: "Initiated", UpdatedAt: timelineNow, Events: []api.TimelineEvent{}},
			{Version: 2, Status: "End", RunState: "Completed", UpdatedAt: timelineNow, Events: []api.TimelineEvent{}},
		},
	}
	require.Equal(t, expected, actual)
}

func TestTimelineHandlerEventWithoutVersion(t *testing.T) {
	listEvents := func(ctx context.Context, runID string) ([]workflow.Event, error) {
		return []workflow.Event{timelineEvent(9, "3")}, nil
	}

	resp := postRun(t, api.Timeline(timelineHistory, listEvents), "1")
	require.Equal(t, http.StatusInternalServerError, resp.Code)
	require.Equal(t, "event 9 has no version in the history of the run (record version \"3\")\n", resp.Body.String())
}
