// Package entities holds the plugin's block types, and nothing else: no
// behaviour, no providers, no clients and no backend libraries.
//
// An application that only reads the items a plugin applied imports this
// package alone, for example xcl.FindByType[entities.Item], so keep it free of
// imports from the rest of the plugin.
package entities

import "github.com/jumppad-labs/xcl/types"

// Item is the block type test "item": a resource that does nothing but record
// each operation xcl asks of it, so a test can read back what ran.
//
//	test "item" "first" {
//	  log    = "./ops.jsonl"
//	  value  = "one"
//	  secret = "hunter2"
//	}
//
//	test "item" "second" {
//	  log   = "./ops.jsonl"
//	  value = "two"
//	  input = test.item.first.output
//	}
type Item struct {
	types.ResourceBase `xcl:",remain"`

	// Log is the path of the JSON-lines file every operation on the item
	// appends its record to, before the operation does its work.
	Log string `xcl:"log" json:"log"`

	// Value is a plain setting. It is updated in place, and output is
	// recomputed from it.
	Value string `xcl:"value,optional" json:"value,omitempty"`

	// Input is usually set from another item's output, which makes this item
	// depend on that one. It is updated in place.
	Input string `xcl:"input,optional" json:"input,omitempty"`

	// Secret is sensitive: xcl redacts it in events and lookups and masks it
	// in state. It is updated in place.
	Secret types.Sensitive[string] `xcl:"secret,optional" json:"secret"`

	// Output is computed by create and update as <name>-<value>.
	Output string `xcl:"output,optional,computed" json:"output,omitempty"`
}
