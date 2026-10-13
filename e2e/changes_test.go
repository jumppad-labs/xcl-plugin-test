package e2e

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jumppad-labs/xcl/diff"
)

// itemAction returns the action the diff plans for the item at address,
// failing the test when the diff does not list it
func itemAction(t *testing.T, changes *diff.Diff, address string) diff.Action {
	t.Helper()

	for _, resource := range changes.Resources {
		if resource.Address == address {
			return resource.Action
		}
	}

	require.Failf(t, "item not in diff", "the diff does not list %s: %+v", address, changes.Resources)

	return ""
}

func TestPlanUpdatesTheItemWhenItsValueChanges(t *testing.T) {
	stateDir := t.TempDir()
	logPath := filepath.Join(t.TempDir(), "ops.jsonl")

	applier := newExternalConfig(t, stateDir, logPath)
	t.Cleanup(func() { require.NoError(t, applier.Destroy()) })
	require.NoError(t, applier.Apply(basicExample))

	planner := newExternalConfig(t, stateDir, logPath)

	changes, err := planner.Diff([]string{writeConfiguration(t, renumberedItems)})
	require.NoError(t, err)
	require.Equal(t, diff.ActionUpdate, itemAction(t, changes, firstAddress))
}
