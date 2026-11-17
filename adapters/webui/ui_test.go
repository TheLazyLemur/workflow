package webui_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luno/workflow"
	"github.com/luno/workflow/adapters/memrecordstore"
	"github.com/stretchr/testify/require"

	"github.com/luno/workflow/adapters/webui"
)

type status int

const (
	statusStart status = 1
	statusEnd   status = 2
)

func (s status) String() string {
	switch s {
	case statusStart:
		return "Start"
	case statusEnd:
		return "End"
	default:
		return "Unknown"
	}
}

func TestGraphHandlerFuncReadsHistory(t *testing.T) {
	b := workflow.NewBuilder[string, status]("test")
	b.AddStep(statusStart, func(ctx context.Context, r *workflow.Run[string, status]) (status, error) {
		return statusEnd, nil
	}, statusEnd)
	wf := b.Build(nil, nil, nil)

	store := memrecordstore.New()
	record := &workflow.Record{WorkflowName: "test", ForeignID: "9", RunID: "1", Status: int(statusStart)}
	err := store.Store(context.Background(), record)
	require.NoError(t, err)

	record.Status = int(statusEnd)
	err = store.Store(context.Background(), record)
	require.NoError(t, err)

	body, err := json.Marshal(map[string]string{"run_id": "1"})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	webui.GraphHandlerFunc(store, wf)(rec, httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body)))
	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Path            []int `json:"path"`
		PathFromHistory bool  `json:"path_from_history"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, []int{int(statusStart), int(statusEnd)}, resp.Path)
	require.True(t, resp.PathFromHistory)
}
