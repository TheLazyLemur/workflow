package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/luno/workflow"
	"github.com/stretchr/testify/require"

	"github.com/luno/workflow/adapters/webui/internal/api"
)

var testGraphs = map[string]workflow.Graph{
	"test": {
		Statuses: []workflow.GraphStatus{
			{Value: 1, Description: "Start", Starting: true},
			{Value: 2, Description: "Middle"},
			{Value: 3, Description: "End", Terminal: true},
		},
		Transitions: []workflow.GraphTransition{
			{From: 1, To: 2},
			{From: 2, To: 3},
		},
	},
}

var expectedGraphStatuses = []api.GraphStatus{
	{Value: 1, Description: "Start", Starting: true},
	{Value: 2, Description: "Middle"},
	{Value: 3, Description: "End", Terminal: true},
}

var expectedGraphTransitions = []api.GraphTransition{
	{From: 1, To: 2},
	{From: 2, To: 3},
}

func lookupRecord(record workflow.Record) api.LookupFn {
	return func(ctx context.Context, runID string) (*workflow.Record, error) {
		return &record, nil
	}
}

func TestGraphHandler(t *testing.T) {
	lookup := lookupRecord(workflow.Record{WorkflowName: "test", RunID: "1", Status: 3})
	listHistory := func(ctx context.Context, runID string) ([]workflow.Record, error) {
		require.Equal(t, "1", runID)
		return []workflow.Record{
			{Status: 1, RunState: workflow.RunStateInitiated},
			{Status: 1, RunState: workflow.RunStateRunning},
			{Status: 2, RunState: workflow.RunStatePaused},
			{Status: 2, RunState: workflow.RunStateRunning},
			{Status: 3, RunState: workflow.RunStateCompleted},
		}, nil
	}

	resp := postRun(t, api.Graph(lookup, listHistory, testGraphs), "1")
	require.Equal(t, http.StatusOK, resp.Code)

	var actual api.GraphResponse
	err := json.Unmarshal(resp.Body.Bytes(), &actual)
	require.NoError(t, err)

	expected := api.GraphResponse{
		Statuses:        expectedGraphStatuses,
		Transitions:     expectedGraphTransitions,
		Path:            []int{1, 2, 3},
		PathFromHistory: true,
	}
	require.Equal(t, expected, actual)
}

func TestGraphHandlerWithoutHistory(t *testing.T) {
	lookup := lookupRecord(workflow.Record{WorkflowName: "test", RunID: "1", Status: 2})

	resp := postRun(t, api.Graph(lookup, nil, testGraphs), "1")
	require.Equal(t, http.StatusOK, resp.Code)

	var actual api.GraphResponse
	err := json.Unmarshal(resp.Body.Bytes(), &actual)
	require.NoError(t, err)

	expected := api.GraphResponse{
		Statuses:        expectedGraphStatuses,
		Transitions:     expectedGraphTransitions,
		Path:            []int{2},
		PathFromHistory: false,
	}
	require.Equal(t, expected, actual)
}

func TestGraphHandlerUnknownWorkflow(t *testing.T) {
	lookup := lookupRecord(workflow.Record{WorkflowName: "other", RunID: "1", Status: 2})

	resp := postRun(t, api.Graph(lookup, nil, testGraphs), "1")
	require.Equal(t, http.StatusNotFound, resp.Code)
	require.Equal(t, "unknown workflow: other\n", resp.Body.String())
}
