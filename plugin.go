// Package testplugin is the test plugin, which provides the block type
// test "item" and records every operation xcl asks of an item to a log the
// configuration chooses.
//
// It exists to test xcl and its language bindings against a real, signed,
// externally released plugin: the GitHub registry's download, checksum and
// signature checks, and the lifecycle, sensitive and computed values through
// a plugin program. It is not an example to copy; see xcl-plugin-template.
//
// Plugin is importable: an application registers it in-process with
// registry.Local.RegisterPlugin(&testplugin.Plugin{}), and
// cmd/xcl-plugin-test serves the same type as a separate program. The block
// type lives in entities, the provider in providers and the backend layer in
// client.
package testplugin

import (
	"github.com/jumppad-labs/xcl/logger"
	"github.com/jumppad-labs/xcl/plugins"

	"github.com/jumppad-labs/xcl-plugin-test/client/oplog"
	"github.com/jumppad-labs/xcl-plugin-test/entities"
	"github.com/jumppad-labs/xcl-plugin-test/providers"
)

// Plugin provides the block type test "item". cmd/xcl-plugin-test serves it
// as a separate plugin program.
type Plugin struct {
	plugins.PluginBase
}

var _ plugins.Plugin = (*Plugin)(nil)

// Init builds the operation log client once, without touching the disk, and
// registers the test "item" block type with a provider built around it.
func (p *Plugin) Init(logger logger.Logger, state plugins.State) error {
	log := oplog.NewFile()

	logger.Debug("registering block types", "block_types", "test.item")

	return plugins.RegisterResourceProvider(
		&p.PluginBase,
		logger,
		state,
		"test",
		"item",
		&entities.Item{},
		providers.NewItemProvider(log),
	)
}
