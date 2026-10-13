package oplog_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jumppad-labs/xcl-plugin-test/client/oplog"
)

func TestAppendThenReadKeepsTheOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ops.jsonl")
	log := oplog.NewFile()

	require.NoError(t, log.Append(context.Background(), path, oplog.Operation{Operation: oplog.Create, Address: "test.item.first"}))
	require.NoError(t, log.Append(context.Background(), path, oplog.Operation{Operation: oplog.Create, Address: "test.item.second"}))
	require.NoError(t, log.Append(context.Background(), path, oplog.Operation{Operation: oplog.Destroy, Address: "test.item.second"}))

	records, err := log.Read(context.Background(), path)
	require.NoError(t, err)

	require.Equal(t, []oplog.Operation{
		{Operation: "create", Address: "test.item.first"},
		{Operation: "create", Address: "test.item.second"},
		{Operation: "destroy", Address: "test.item.second"},
	}, records)
}

func TestAppendWritesOneJSONLinePerRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ops.jsonl")
	log := oplog.NewFile()

	require.NoError(t, log.Append(context.Background(), path, oplog.Operation{Operation: oplog.Update, Address: "test.item.first"}))

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, `{"operation":"update","address":"test.item.first"}`+"\n", string(content))
}

func TestReadOfAMissingLogIsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.jsonl")

	records, err := oplog.NewFile().Read(context.Background(), path)

	require.NoError(t, err)
	require.Empty(t, records)
}

func TestAppendToAnUnwritablePathReturnsAnError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no-such-directory", "ops.jsonl")

	err := oplog.NewFile().Append(context.Background(), path, oplog.Operation{Operation: oplog.Create, Address: "test.item.first"})

	require.Error(t, err)
	require.Contains(t, err.Error(), path)
}

func TestReadOfALineThatIsNotJSONReturnsAnError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ops.jsonl")
	require.NoError(t, os.WriteFile(path, []byte("not json\n"), 0o644))

	records, err := oplog.NewFile().Read(context.Background(), path)

	require.Error(t, err)
	require.Nil(t, records)
}

func TestAppendWithACancelledContextReturnsContextCanceledAndWritesNothing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ops.jsonl")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := oplog.NewFile().Append(ctx, path, oplog.Operation{Operation: oplog.Create, Address: "test.item.first"})
	require.ErrorIs(t, err, context.Canceled)

	_, err = os.Stat(path)
	require.ErrorIs(t, err, os.ErrNotExist)
}
