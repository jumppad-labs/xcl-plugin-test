// Package providers holds the plugin's resource providers, one file per
// resource type.
//
// Each provider is built by a constructor that takes the clients it uses, so
// the plugin passes it the real clients and the unit tests pass strict
// doubles. Methods follow the lifecycle order: Init, Create, Read, Changed,
// Update, Destroy, Functions.
package providers

import (
	"context"
	"fmt"

	"github.com/jumppad-labs/xcl/entity"
	"github.com/jumppad-labs/xcl/logger"
	"github.com/jumppad-labs/xcl/plugins"

	"github.com/jumppad-labs/xcl-plugin-test/client/oplog"
	"github.com/jumppad-labs/xcl-plugin-test/entities"
)

// ItemProvider records every operation on a test "item" block to the item's
// log, through the operation log it was given, before doing the operation's
// work, which is only to compute output. Changed is DefaultChanged's, so
// every setting is updated in place and an item is never replaced.
type ItemProvider struct {
	plugins.DefaultChanged[*entities.Item]

	log oplog.OperationLog
}

var _ plugins.ResourceProvider[*entities.Item] = (*ItemProvider)(nil)

// NewItemProvider returns the provider for test "item" blocks, which records
// its operations only through log.
func NewItemProvider(log oplog.OperationLog) *ItemProvider {
	return &ItemProvider{log: log}
}

// Init logs that the provider is ready, it needs nothing else.
func (p *ItemProvider) Init(state plugins.State, functions plugins.ProviderFunctions, log logger.Logger) error {
	log.Debug("provider ready")
	return nil
}

// Create records the create, then computes output.
func (p *ItemProvider) Create(ctx context.Context, item *entities.Item) (*entities.Item, error) {
	if err := p.record(ctx, oplog.Create, item); err != nil {
		return nil, err
	}

	item.Output = output(item)
	plugins.Logger(ctx).Info("created item", "address", item.Meta.ID, "output", item.Output)

	return item, nil
}

// Read records the read and keeps the output the item was created or last
// updated with, since nothing outside the log holds an item.
func (p *ItemProvider) Read(ctx context.Context, old *entities.Item, new *entities.Item) (*entities.Item, error) {
	if err := p.record(ctx, oplog.Read, new); err != nil {
		return nil, err
	}

	new.Output = old.Output

	return new, nil
}

// Update records the update, then computes output again from the item's
// value.
func (p *ItemProvider) Update(ctx context.Context, item *entities.Item, changes []entity.PropertyChange, dependencies []entity.DependencyChange) (*entities.Item, error) {
	if err := p.record(ctx, oplog.Update, item); err != nil {
		return nil, err
	}

	item.Output = output(item)
	plugins.Logger(ctx).Info("updated item", "address", item.Meta.ID, "output", item.Output, "changes", len(changes))

	return item, nil
}

// Destroy records the destroy. The record outlives the item, so a test can
// read destroys back.
func (p *ItemProvider) Destroy(ctx context.Context, item *entities.Item, force bool) error {
	if err := p.record(ctx, oplog.Destroy, item); err != nil {
		return err
	}

	plugins.Logger(ctx).Info("destroyed item", "address", item.Meta.ID)

	return nil
}

// Functions returns nil: the provider offers no functions to configuration.
func (p *ItemProvider) Functions() plugins.ProviderFunctions {
	return nil
}

// record appends the record of operation on item to the item's log.
func (p *ItemProvider) record(ctx context.Context, operation string, item *entities.Item) error {
	err := p.log.Append(ctx, item.Log, oplog.Operation{Operation: operation, Address: item.Meta.ID})
	if err != nil {
		return fmt.Errorf("unable to record the %s of %s: %w", operation, item.Meta.ID, err)
	}

	return nil
}

// output is an item's computed output, <name>-<value>.
func output(item *entities.Item) string {
	return item.Meta.Name + "-" + item.Value
}
