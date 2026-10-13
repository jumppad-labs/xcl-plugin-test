// Command xcl-plugin-test serves the test plugin as a separate plugin program.
// xcl starts this binary and talks to it over gRPC; the plugin it serves is
// the same testplugin.Plugin an application can register in-process.
//
// Build it with `make build`, which writes build/xcl-plugin-test.
package main

import (
	"github.com/jumppad-labs/xcl/plugins"

	testplugin "github.com/jumppad-labs/xcl-plugin-test"
)

// main serves the plugin to the host that started this process
func main() {
	plugins.Serve(&testplugin.Plugin{})
}
