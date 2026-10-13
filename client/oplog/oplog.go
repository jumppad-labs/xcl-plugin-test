// Package oplog is the plugin's only backend layer: the log of operations an
// item records to.
//
// The item provider never touches the filesystem itself. It holds an
// OperationLog, a narrow interface with only the calls it makes, so its unit
// tests replace it with the strict double Mockery generates into
// client/oplog/mocks. File is the real implementation, an append-only file of
// JSON lines that a test in any language can read back.
package oplog

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
)

// Operation is one line of the log: the operation that started and the
// address of the item it started for. It is written as
// {"operation":"create","address":"test.item.first"}.
type Operation struct {
	// Operation is create, read, update or destroy.
	Operation string `json:"operation"`
	// Address is the item's address, such as test.item.first.
	Address string `json:"address"`
}

// The operations an item records.
const (
	Create  = "create"
	Read    = "read"
	Update  = "update"
	Destroy = "destroy"
)

// OperationLog is the log the item provider records to. Every method takes the
// context of the call it serves and returns an error once it is cancelled.
type OperationLog interface {
	// Append adds record to the end of the log at path, creating the file
	// when it does not exist.
	Append(ctx context.Context, path string, record Operation) error
	// Read returns every record in the log at path, in the order they were
	// appended. A log that does not exist holds no records.
	Read(ctx context.Context, path string) ([]Operation, error)
}

// File is the OperationLog over files on the local disk.
type File struct {
	// mu keeps the lines of concurrent appends whole and in order
	mu sync.Mutex
}

var _ OperationLog = (*File)(nil)

// NewFile returns an OperationLog over files on the local disk. It touches
// nothing until one of its methods is called, so a plugin can build it in
// Init.
func NewFile() *File {
	return &File{}
}

// Append adds record as one JSON line to the end of the file at path.
func (f *File) Append(ctx context.Context, path string, record Operation) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	line, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encoding the %s record of %s: %w", record.Operation, record.Address, err)
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("opening the operation log %s: %w", path, err)
	}

	_, writeErr := file.Write(append(line, '\n'))
	closeErr := file.Close()

	if writeErr != nil {
		return fmt.Errorf("writing to the operation log %s: %w", path, writeErr)
	}

	if closeErr != nil {
		return fmt.Errorf("closing the operation log %s: %w", path, closeErr)
	}

	return nil
}

// Read returns the records in the file at path, in order. A file that does
// not exist holds no records.
func (f *File) Read(ctx context.Context, path string) ([]Operation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return []Operation{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("reading the operation log %s: %w", path, err)
	}

	records := []Operation{}

	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		if len(bytes.TrimSpace(scanner.Bytes())) == 0 {
			continue
		}

		var record Operation
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, fmt.Errorf("decoding a line of the operation log %s: %w", path, err)
		}

		records = append(records, record)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading the operation log %s: %w", path, err)
	}

	return records, nil
}
