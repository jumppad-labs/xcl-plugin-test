// Package stateonly reads what the test plugin applied as an application
// that imports only the entities package would: it never imports the plugin,
// its providers or its clients, only the block type.
package stateonly

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jumppad-labs/xcl"
	"github.com/jumppad-labs/xcl/registry"

	"github.com/jumppad-labs/xcl-plugin-test/entities"
)

// basicExample is the directory holding the sample configuration the test
// applies
const basicExample = "../../examples/basic"

// pluginBinary is the path of the plugin program TestMain builds for the
// tests in this package
var pluginBinary string

func TestMain(m *testing.M) {
	buildDir, err := os.MkdirTemp("", "xcl-plugin-test-stateonly")
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create build directory: %s\n", err)
		os.Exit(1)
	}

	pluginBinary = filepath.Join(buildDir, "xcl-plugin-test")

	build := exec.Command("go", "build", "-o", pluginBinary, "../../cmd/xcl-plugin-test")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr

	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "unable to build the plugin program: %s\n", err)
		os.RemoveAll(buildDir)
		os.Exit(1)
	}

	code := m.Run()

	os.RemoveAll(buildDir)
	os.Exit(code)
}

// newConfig returns a Config with the plugin program TestMain built
// registered as an external plugin, keeping its state in stateDir and
// recording operations to logPath
func newConfig(t *testing.T, stateDir string, logPath string) *xcl.Config {
	t.Helper()

	local := registry.NewLocal()
	local.RegisterExternalPlugin(pluginBinary)

	c, err := xcl.NewConfig(
		xcl.WithRegistry(local),
		xcl.WithStatePath(stateDir),
		xcl.WithVariables(map[string]any{"log": logPath}),
	)
	require.NoError(t, err)

	return c
}

func TestStateReaderReadsAppliedItemsIntoTheEntityType(t *testing.T) {
	stateDir := t.TempDir()
	logPath := filepath.Join(t.TempDir(), "ops.jsonl")

	applier := newConfig(t, stateDir, logPath)

	// register the destroy before applying, so a partly applied
	// configuration is removed too
	t.Cleanup(func() {
		require.NoError(t, applier.Destroy())
	})

	err := applier.Apply(basicExample)
	require.NoError(t, err)

	// a separate Config on the same state, as another program reading what
	// was applied would be
	reader := newConfig(t, stateDir, logPath)

	err = reader.Load()
	require.NoError(t, err)

	found, err := xcl.FindByType[entities.Item](reader, "test", "item")
	require.NoError(t, err)
	require.Len(t, found, 2)

	outputs := map[string]string{}
	for _, item := range found {
		outputs[item.Meta.Name] = item.Output
	}

	require.Equal(t, map[string]string{"first": "first-one", "second": "second-two"}, outputs)
}
