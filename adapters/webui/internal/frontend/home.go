package frontend

import (
	_ "embed"
	"net/http"
	"text/template"
)

//go:embed home.html
var homeTemplate string

func HomeHandlerFunc(paths Paths) http.HandlerFunc {
	// Use [[ ]] delimiters so JSX {{ }} syntax passes through unmodified.
	t := template.Must(template.New("home.html").Delims("[[", "]]").Parse(homeTemplate))
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := t.Execute(w, paths); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type Paths struct {
	List       string
	Update     string
	ObjectData string

	// Graph is the path of the handler returned by GraphHandlerFunc. The run's graph is only shown when it is set.
	Graph string
	// Steps is the path of the handler returned by StepsHandlerFunc. The run's steps are only shown when it is set.
	Steps string
	// Timeline is the path of the handler returned by TimelineHandlerFunc. When it is not set the run's timeline is
	// drawn from its current run state only.
	Timeline string
}
