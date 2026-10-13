// Package e2e runs the test plugin end to end: it applies, plans, updates and
// destroys the sample configuration in examples/basic, with the plugin
// registered in-process and with the plugin built and run as a separate
// plugin program, and reads the operation log back.
package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jumppad-labs/xcl"
	"github.com/jumppad-labs/xcl/registry"

	testplugin "github.com/jumppad-labs/xcl-plugin-test"
	"github.com/jumppad-labs/xcl-plugin-test/client/oplog"
)

// basicExample is the directory holding the sample configuration the tests
// apply
const basicExample = "../examples/basic"

// the addresses of the two items the sample configuration declares; second
// takes first's output as its input, so it depends on first
const (
	firstAddress  = "test.item.first"
	secondAddress = "test.item.second"
)

// pluginBinary is the path of the plugin program TestMain builds for the
// tests in this package
var pluginBinary string

func TestMain(m *testing.M) {
	buildDir, err := os.MkdirTemp("", "xcl-plugin-test-e2e")
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create build directory: %s\n", err)
		os.Exit(1)
	}

	pluginBinary = filepath.Join(buildDir, "xcl-plugin-test")

	build := exec.Command("go", "build", "-o", pluginBinary, "../cmd/xcl-plugin-test")
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

// newInProcessConfig returns a Config with the plugin registered in-process,
// keeping its state in stateDir and recording operations to logPath
func newInProcessConfig(t *testing.T, stateDir string, logPath string, options ...xcl.ConfigOption) *xcl.Config {
	t.Helper()

	local := registry.NewLocal()
	local.RegisterPlugin(&testplugin.Plugin{})

	return newConfig(t, local, stateDir, logPath, options...)
}

// newExternalConfig returns a Config with the plugin program TestMain built
// registered as an external plugin, keeping its state in stateDir and
// recording operations to logPath
func newExternalConfig(t *testing.T, stateDir string, logPath string, options ...xcl.ConfigOption) *xcl.Config {
	t.Helper()

	local := registry.NewLocal()
	local.RegisterExternalPlugin(pluginBinary)

	return newConfig(t, local, stateDir, logPath, options...)
}

func newConfig(t *testing.T, local *registry.Local, stateDir string, logPath string, options ...xcl.ConfigOption) *xcl.Config {
	t.Helper()

	all := append([]xcl.ConfigOption{
		xcl.WithRegistry(local),
		xcl.WithStatePath(stateDir),
		xcl.WithVariables(map[string]any{"log": logPath}),
	}, options...)

	c, err := xcl.NewConfig(all...)
	require.NoError(t, err)

	return c
}

// readLog reads the operation log at logPath back
func readLog(t *testing.T, logPath string) []oplog.Operation {
	t.Helper()

	records, err := oplog.NewFile().Read(context.Background(), logPath)
	require.NoError(t, err)

	return records
}

// position returns the index of the first record of operation on address in
// records, failing the test when there is none
func position(t *testing.T, records []oplog.Operation, operation string, address string) int {
	t.Helper()

	for i, record := range records {
		if record.Operation == operation && record.Address == address {
			return i
		}
	}

	require.Failf(t, "operation not recorded", "no %s of %s in %+v", operation, address, records)

	return -1
}

// operationsOf returns the operations recorded in records, in order, without
// the reads, which xcl makes as often as it needs to
func operationsOf(records []oplog.Operation, operation string) []oplog.Operation {
	found := []oplog.Operation{}

	for _, record := range records {
		if record.Operation == operation {
			found = append(found, record)
		}
	}

	return found
}

// recordOf is the record of operation on address
func recordOf(operation string, address string) oplog.Operation {
	return oplog.Operation{Operation: operation, Address: address}
}
