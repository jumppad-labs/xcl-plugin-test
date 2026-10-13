package testplugin

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jumppad-labs/xcl/logger"
)

func TestInitRegistersTheItemType(t *testing.T) {
	directory := t.TempDir()
	t.Chdir(directory)

	p := &Plugin{}

	err := p.Init(logger.Nop(), nil)
	require.NoError(t, err)

	registered := p.GetTypes()
	require.Len(t, registered, 1)
	require.Equal(t, "test", registered[0].Type)
	require.Equal(t, "item", registered[0].SubType)

	entries, err := os.ReadDir(directory)
	require.NoError(t, err)
	require.Empty(t, entries)
}
