package workflow_test

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luno/workflow"
)

func TestGraph(t *testing.T) {
	b := workflow.NewBuilder[string, status]("example")
	b.AddStep(StatusStart, func(ctx context.Context, r *workflow.Run[string, status]) (status, error) {
		return StatusMiddle, nil
	}, StatusMiddle, StatusEnd)
	b.AddStep(StatusMiddle, func(ctx context.Context, r *workflow.Run[string, status]) (status, error) {
		return StatusEnd, nil
	}, StatusEnd)
	b.AddCallback(StatusInitiated, func(ctx context.Context, r *workflow.Run[string, status], reader io.Reader) (status, error) {
		return StatusStart, nil
	}, StatusStart)

	wf := b.Build(nil, nil, nil)

	expected := workflow.Graph{
		Statuses: []workflow.GraphStatus{
			{Value: int(StatusInitiated), Description: "Initiated", Starting: true},
			{Value: int(StatusStart), Description: "Start"},
			{Value: int(StatusMiddle), Description: "Middle"},
			{Value: int(StatusEnd), Description: "End", Terminal: true},
		},
		Transitions: []workflow.GraphTransition{
			{From: int(StatusStart), To: int(StatusMiddle)},
			{From: int(StatusStart), To: int(StatusEnd)},
			{From: int(StatusMiddle), To: int(StatusEnd)},
			{From: int(StatusInitiated), To: int(StatusStart)},
		},
	}
	require.Equal(t, expected, wf.Graph())
}
