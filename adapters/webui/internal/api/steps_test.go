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

func TestStepsHandler(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	listHistory := func(ctx context.Context, runID string) ([]workflow.Record, error) {
		require.Equal(t, "1", runID)
		return []workflow.Record{
			{
				RunState:  workflow.RunStateInitiated,
				Object:    []byte(`{"name":"a"}`),
				UpdatedAt: now,
				Meta:      workflow.Meta{Version: 1, StatusDescription: "Start"},
			},
			{
				RunState:  workflow.RunStateCompleted,
				Object:    []byte(`{"name":"b"}`),
				UpdatedAt: now.Add(time.Second),
				Meta:      workflow.Meta{Version: 2, StatusDescription: "End"},
			},
			{
				RunState:  workflow.RunStateDataDeleted,
				Object:    []byte(`not json`),
				UpdatedAt: now.Add(2 * time.Second),
				Meta:      workflow.Meta{Version: 3, StatusDescription: "End"},
			},
		}, nil
	}

	resp := postRun(t, api.Steps(listHistory), "1")
	require.Equal(t, http.StatusOK, resp.Code)

	var actual api.StepsResponse
	err := json.Unmarshal(resp.Body.Bytes(), &actual)
	require.NoError(t, err)

	expected := api.StepsResponse{
		Items: []api.StepItem{
			{Version: 1, Status: "Start", RunState: "Initiated", UpdatedAt: now, Object: json.RawMessage(`{"name":"a"}`)},
			{Version: 2, Status: "End", RunState: "Completed", UpdatedAt: now.Add(time.Second), Object: json.RawMessage(`{"name":"b"}`)},
			{Version: 3, Status: "End", RunState: "Data Deleted", UpdatedAt: now.Add(2 * time.Second), Object: json.RawMessage(`"not json"`)},
		},
	}
	require.Equal(t, expected, actual)
}

func TestStepsHandlerHistoryError(t *testing.T) {
	listHistory := func(ctx context.Context, runID string) ([]workflow.Record, error) {
		return nil, workflow.ErrRecordNotFound
	}

	resp := postRun(t, api.Steps(listHistory), "1")
	require.Equal(t, http.StatusInternalServerError, resp.Code)
}
