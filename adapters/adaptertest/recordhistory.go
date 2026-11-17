package adaptertest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luno/workflow"
)

// RunRecordHistoryTest tests a RecordStore that implements the optional workflow.RecordHistory interface. It should
// be run in addition to RunRecordStoreTest.
func RunRecordHistoryTest(t *testing.T, factory func() workflow.RecordStore) {
	tests := []func(t *testing.T, factory func() workflow.RecordStore){
		testListHistory,
		testListHistoryOnlyIncludesRun,
		testListHistoryNotFound,
		testListHistoryAfterDataDeleted,
	}

	for _, test := range tests {
		test(t, factory)
	}
}

func testListHistory(t *testing.T, factory func() workflow.RecordStore) {
	t.Run("ListHistory", func(t *testing.T) {
		store := factory()
		ctx := context.Background()
		record := dummyWireRecord(t, "my_workflow")

		expected := []workflow.Record{
			storeVersion(t, store, record, statusStarted, workflow.RunStateInitiated),
			storeVersion(t, store, record, statusMiddle, workflow.RunStateRunning),
			storeVersion(t, store, record, statusEnd, workflow.RunStateCompleted),
		}

		actual, err := recordHistory(t, store).ListHistory(ctx, record.RunID)
		require.NoError(t, err)

		require.Equal(t, len(expected), len(actual))
		for i := range expected {
			recordIsEqual(t, expected[i], actual[i])
			require.Equal(t, expected[i].Meta.Version, actual[i].Meta.Version)
			require.Equal(t, expected[i].Meta.StatusDescription, actual[i].Meta.StatusDescription)
		}
	})
}

func testListHistoryOnlyIncludesRun(t *testing.T, factory func() workflow.RecordStore) {
	t.Run("ListHistory only includes the run", func(t *testing.T) {
		store := factory()
		ctx := context.Background()
		other := dummyWireRecord(t, "my_workflow")
		record := dummyWireRecord(t, "my_workflow")

		storeVersion(t, store, other, statusStarted, workflow.RunStateInitiated)
		storeVersion(t, store, other, statusMiddle, workflow.RunStateRunning)
		expected := storeVersion(t, store, record, statusStarted, workflow.RunStateInitiated)

		actual, err := recordHistory(t, store).ListHistory(ctx, record.RunID)
		require.NoError(t, err)

		require.Equal(t, 1, len(actual))
		recordIsEqual(t, expected, actual[0])
	})
}

func testListHistoryNotFound(t *testing.T, factory func() workflow.RecordStore) {
	t.Run("ListHistory returns ErrRecordNotFound", func(t *testing.T) {
		store := factory()

		_, err := recordHistory(t, store).ListHistory(context.Background(), "unknown-run-id")
		require.ErrorIs(t, err, workflow.ErrRecordNotFound)
	})
}

func testListHistoryAfterDataDeleted(t *testing.T, factory func() workflow.RecordStore) {
	t.Run("ListHistory does not return deleted data", func(t *testing.T) {
		store := factory()
		ctx := context.Background()
		record := dummyWireRecord(t, "my_workflow")

		storeVersion(t, store, record, statusStarted, workflow.RunStateInitiated)
		storeVersion(t, store, record, statusEnd, workflow.RunStateCompleted)

		replacement := []byte(`{"result":"deleted"}`)
		record.Object = replacement
		storeVersion(t, store, record, statusEnd, workflow.RunStateDataDeleted)

		actual, err := recordHistory(t, store).ListHistory(ctx, record.RunID)
		require.NoError(t, err)

		require.Equal(t, 3, len(actual))
		for _, version := range actual {
			require.Equal(t, replacement, version.Object)
		}
	})
}

func recordHistory(t *testing.T, store workflow.RecordStore) workflow.RecordHistory {
	history, ok := store.(workflow.RecordHistory)
	require.True(t, ok, "%T does not implement workflow.RecordHistory", store)
	return history
}

// storeVersion stores the next version of the record. The same record is updated and stored again for each version,
// as workflow.RunStateController does, so stores must not keep a reference to it.
func storeVersion(t *testing.T, store workflow.RecordStore, record *workflow.Record, s status, runState workflow.RunState) workflow.Record {
	record.Status = int(s)
	record.RunState = runState
	record.Meta.StatusDescription = s.String()
	record.Meta.Version++

	err := store.Store(context.Background(), record)
	require.NoError(t, err)

	return *record
}
