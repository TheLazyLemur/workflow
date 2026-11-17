package workflow

import "slices"

// Graph describes the statuses of a workflow and the transitions between them as they were declared on the Builder.
// It holds no generic types so that it can be used without knowing the workflow's Type and Status, such as when
// rendering the workflow in a user interface.
type Graph struct {
	// Statuses holds every status in the graph ordered by value.
	Statuses []GraphStatus

	// Transitions holds every transition in the graph. Transitions from the same status are grouped together and
	// appear in the order they were added to the Builder.
	Transitions []GraphTransition
}

// GraphStatus is a single status in a Graph.
type GraphStatus struct {
	// Value is the int value of the status.
	Value int

	// Description is the String value of the status.
	Description string

	// Starting is true when no transition leads to the status, which makes it a status that a Run can be
	// triggered at.
	Starting bool

	// Terminal is true when no transition leads away from the status. A Run that reaches a terminal status is
	// completed.
	Terminal bool
}

// GraphTransition is a transition between two statuses in a Graph.
type GraphTransition struct {
	From int
	To   int

	// Step is the name, given with WithName, of the step that consumes From and can move a Run to To. It is empty
	// when there is no such step or it has no name.
	Step string
}

// Graph returns the statuses of the workflow and the transitions between them.
func (w *Workflow[Type, Status]) Graph() Graph {
	info := w.statusGraph.Info()

	var g Graph
	for _, node := range w.statusGraph.Nodes() {
		g.Statuses = append(g.Statuses, GraphStatus{
			Value:       node,
			Description: Status(node).String(),
			Starting:    slices.Contains(info.StartingNodes, node),
			Terminal:    w.statusGraph.IsTerminal(node),
		})
	}

	for _, transition := range info.Transitions {
		var step string
		consumer, ok := w.consumers[Status(transition.From)]
		if ok && slices.Contains(consumer.destinations, Status(transition.To)) {
			step = consumer.name
		}

		g.Transitions = append(g.Transitions, GraphTransition{
			From: transition.From,
			To:   transition.To,
			Step: step,
		})
	}

	return g
}
