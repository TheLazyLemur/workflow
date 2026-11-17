package api

import (
	"net/http"

	"github.com/luno/workflow"
)

type GraphResponse struct {
	Statuses    []GraphStatus     `json:"statuses"`
	Transitions []GraphTransition `json:"transitions"`

	// Path holds the statuses the run has been in, oldest first, with repeats of the same status collapsed.
	Path []int `json:"path"`
	// PathFromHistory is false when the record store does not keep history, in which case Path only holds the run's
	// current status.
	PathFromHistory bool `json:"path_from_history"`
}

type GraphStatus struct {
	Value       int    `json:"value"`
	Description string `json:"description"`
	Starting    bool   `json:"starting"`
	Terminal    bool   `json:"terminal"`
}

type GraphTransition struct {
	From int `json:"from"`
	To   int `json:"to"`
}

// Graph serves the graph of the run's workflow from graphs, keyed by workflow name. listHistory is optional.
func Graph(lookup LookupFn, listHistory ListHistoryFn, graphs map[string]workflow.Graph) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := readRunRequest(w, r)
		if !ok {
			return
		}

		record, err := lookup(r.Context(), req.RunID)
		if err != nil {
			http.Error(w, "failed to lookup record from store", http.StatusInternalServerError)
			return
		}

		graph, ok := graphs[record.WorkflowName]
		if !ok {
			http.Error(w, "unknown workflow: "+record.WorkflowName, http.StatusNotFound)
			return
		}

		resp := GraphResponse{
			Statuses:    []GraphStatus{},
			Transitions: []GraphTransition{},
			Path:        []int{record.Status},
		}

		if listHistory != nil {
			history, err := listHistory(r.Context(), req.RunID)
			if err != nil {
				http.Error(w, "failed to list history from store", http.StatusInternalServerError)
				return
			}

			resp.Path = []int{}
			for _, version := range history {
				if len(resp.Path) > 0 && resp.Path[len(resp.Path)-1] == version.Status {
					continue
				}

				resp.Path = append(resp.Path, version.Status)
			}

			resp.PathFromHistory = true
		}

		for _, status := range graph.Statuses {
			resp.Statuses = append(resp.Statuses, GraphStatus{
				Value:       status.Value,
				Description: status.Description,
				Starting:    status.Starting,
				Terminal:    status.Terminal,
			})
		}

		for _, transition := range graph.Transitions {
			resp.Transitions = append(resp.Transitions, GraphTransition{
				From: transition.From,
				To:   transition.To,
			})
		}

		writeJSON(w, resp)
	}
}
