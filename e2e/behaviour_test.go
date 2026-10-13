package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jumppad-labs/xcl"
	"github.com/jumppad-labs/xcl/types"

	"github.com/jumppad-labs/xcl-plugin-test/entities"
)

// The behaviours below are checked once with the plugin in-process and once
// as a separate program, by inprocess_test.go and external_test.go.

// configFactory builds a Config around the plugin, keeping its state in
// stateDir and recording operations to logPath
type configFactory func(t *testing.T, stateDir string, logPath string, options ...xcl.ConfigOption) *xcl.Config

// renumberedItems is the sample configuration with first's value changed,
// which changes first's output and so second's input
const renumberedItems = `
variable "log" {
  default = "./ops.jsonl"
}

test "item" "first" {
  log    = variable.log
  value  = "uno"
  secret = "hunter2"
}

test "item" "second" {
  log   = variable.log
  value = "two"
  input = test.item.first.output
}
`

// writeConfiguration writes content as main.xcl in a new temporary directory
// and returns the directory
func writeConfiguration(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "main.xcl"), []byte(content), 0o644)
	require.NoError(t, err)

	return dir
}

func appliesPlansNoChangesAndDestroys(t *testing.T, newConfig configFactory) {
	logPath := filepath.Join(t.TempDir(), "ops.jsonl")
	c := newConfig(t, t.TempDir(), logPath)

	err := c.Apply(basicExample)
	require.NoError(t, err)

	items, err := xcl.FindByType[entities.Item](c, "test", "item")
	require.NoError(t, err)
	require.Len(t, items, 2)

	changes, err := c.Diff([]string{basicExample})
	require.NoError(t, err)
	require.Equal(t, 0, changes.Changed())

	err = c.Destroy()
	require.NoError(t, err)

	remaining, err := xcl.FindByType[entities.Item](c, "test", "item")
	require.NoError(t, err)
	require.Empty(t, remaining)
}

func recordsCreatesAndDestroysInDependencyOrder(t *testing.T, newConfig configFactory) {
	logPath := filepath.Join(t.TempDir(), "ops.jsonl")
	c := newConfig(t, t.TempDir(), logPath)

	require.NoError(t, c.Apply(basicExample))
	require.NoError(t, c.Destroy())

	records := readLog(t, logPath)

	require.Len(t, operationsOf(records, "create"), 2)
	require.Len(t, operationsOf(records, "destroy"), 2)
	require.Less(t, position(t, records, "create", firstAddress), position(t, records, "create", secondAddress))
	require.Less(t, position(t, records, "destroy", secondAddress), position(t, records, "destroy", firstAddress))
}

func computesOutputOnCreate(t *testing.T, newConfig configFactory) {
	logPath := filepath.Join(t.TempDir(), "ops.jsonl")
	c := newConfig(t, t.TempDir(), logPath)
	t.Cleanup(func() { require.NoError(t, c.Destroy()) })

	require.NoError(t, c.Apply(basicExample))

	first, err := xcl.Find[entities.Item](c, firstAddress)
	require.NoError(t, err)
	require.Equal(t, "first-one", first.Output)

	second, err := xcl.Find[entities.Item](c, secondAddress)
	require.NoError(t, err)
	require.Equal(t, "second-two", second.Output)
	require.Equal(t, "first-one", second.Input)
}

func updatesValueInPlaceAndRecomputesOutput(t *testing.T, newConfig configFactory) {
	stateDir := t.TempDir()
	logPath := filepath.Join(t.TempDir(), "ops.jsonl")

	c := newConfig(t, stateDir, logPath)
	t.Cleanup(func() { require.NoError(t, c.Destroy()) })

	require.NoError(t, c.Apply(basicExample))
	require.NoError(t, c.Apply(writeConfiguration(t, renumberedItems)))

	first, err := xcl.Find[entities.Item](c, firstAddress)
	require.NoError(t, err)
	require.Equal(t, "first-uno", first.Output)

	second, err := xcl.Find[entities.Item](c, secondAddress)
	require.NoError(t, err)
	require.Equal(t, "first-uno", second.Input)

	records := readLog(t, logPath)
	require.Contains(t, operationsOf(records, "update"), recordOf("update", firstAddress))
	require.Empty(t, operationsOf(records, "destroy"))
}

func treatsTheSecretAsSensitive(t *testing.T, newConfig configFactory) {
	var mu sync.Mutex
	var data []string

	handler := func(event xcl.Event) {
		if len(event.Data) == 0 {
			return
		}

		mu.Lock()
		defer mu.Unlock()
		data = append(data, string(event.Data))
	}

	logPath := filepath.Join(t.TempDir(), "ops.jsonl")
	c := newConfig(t, t.TempDir(), logPath, xcl.WithEventHandler(handler), xcl.WithEventData(xcl.EventDataProcessed))
	t.Cleanup(func() { require.NoError(t, c.Destroy()) })

	require.NoError(t, c.Apply(basicExample))

	first, err := xcl.Find[entities.Item](c, firstAddress)
	require.NoError(t, err)
	require.Equal(t, "hunter2", first.Secret.Reveal())

	mu.Lock()
	defer mu.Unlock()

	require.NotEmpty(t, data)

	masked := false
	for _, carried := range data {
		require.NotContains(t, carried, "hunter2")

		if strings.Contains(carried, types.SensitiveMarker) {
			masked = true
		}
	}

	require.True(t, masked, "no event carried the masked secret: %v", data)
}
