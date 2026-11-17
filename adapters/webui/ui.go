package webui

import (
	"fmt"
	"net/http"

	"github.com/luno/workflow"

	"github.com/luno/workflow/adapters/webui/internal/api"
	"github.com/luno/workflow/adapters/webui/internal/frontend"
)

func HomeHandlerFunc(paths Paths) http.HandlerFunc {
	return frontend.HomeHandlerFunc(paths)
}

type (
	Stringer            = api.Stringer
	ListWorkflowRecords = api.ListWorkflowRecords
	LookupFn            = api.LookupFn
	Paths               = frontend.Paths

	// ListRunEvents returns the events that the workflow published for a run, such as those read from the
	// workflow's topics with a run_id header that matches the run.
	ListRunEvents = api.ListRunEvents
)

// GraphSource provides the status graph of a workflow. *workflow.Workflow implements it.
type GraphSource interface {
	Name() string
	Graph() workflow.Graph
}

func ListHandlerFunc(store workflow.RecordStore) http.HandlerFunc {
	return api.List(store.List)
}

func ObjectDataHandlerFunc(store workflow.RecordStore) http.HandlerFunc {
	return api.ObjectData(store.Lookup)
}

func UpdateHandlerFunc(store workflow.RecordStore) http.HandlerFunc {
	return api.Update(store)
}

// GraphHandlerFunc serves the status graph of a run's workflow and the statuses that the run has been in. The
// statuses are read from the store's history when it implements workflow.RecordHistory. Otherwise only the run's
// current status is known. Runs of workflows that are not provided get a 404. Serve it at Paths.Graph.
func GraphHandlerFunc(store workflow.RecordStore, workflows ...GraphSource) http.HandlerFunc {
	graphs := make(map[string]workflow.Graph)
	for _, w := range workflows {
		if _, ok := graphs[w.Name()]; ok {
			panic(fmt.Sprintf("webui: workflow %q provided more than once", w.Name()))
		}

		graphs[w.Name()] = w.Graph()
	}

	var listHistory api.ListHistoryFn
	if history, ok := store.(workflow.RecordHistory); ok {
		listHistory = history.ListHistory
	}

	return api.Graph(store.Lookup, listHistory, graphs)
}

// StepsHandlerFunc serves every saved version of a run, oldest first, with its object so that the web UI can show
// what each step changed. Serve it at Paths.Steps.
func StepsHandlerFunc(history workflow.RecordHistory) http.HandlerFunc {
	return api.Steps(history.ListHistory)
}

// TimelineHandlerFunc serves every saved version of a run, oldest first. listEvents is optional: when it is not nil
// each version lists the events that were published for it. Serve it at Paths.Timeline.
func TimelineHandlerFunc(history workflow.RecordHistory, listEvents ListRunEvents) http.HandlerFunc {
	return api.Timeline(history.ListHistory, listEvents)
}
