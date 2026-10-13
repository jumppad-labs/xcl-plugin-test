package providers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jumppad-labs/xcl/entity"
	"github.com/jumppad-labs/xcl/types"

	"github.com/jumppad-labs/xcl-plugin-test/client/oplog"
	"github.com/jumppad-labs/xcl-plugin-test/client/oplog/mocks"
	"github.com/jumppad-labs/xcl-plugin-test/entities"
	"github.com/jumppad-labs/xcl-plugin-test/providers"
)

func testItem() *entities.Item {
	return &entities.Item{
		ResourceBase: types.ResourceBase{
			Meta: types.Meta{ID: "test.item.first", Name: "first", Type: "test"},
		},
		Log:   "ops.jsonl",
		Value: "one",
	}
}

func expectAppend(log *mocks.MockOperationLog, operation string) {
	log.EXPECT().
		Append(mock.Anything, "ops.jsonl", oplog.Operation{Operation: operation, Address: "test.item.first"}).
		Return(nil).
		Once()
}

func TestCreateRecordsTheCreate(t *testing.T) {
	log := mocks.NewMockOperationLog(t)
	expectAppend(log, "create")

	_, err := providers.NewItemProvider(log).Create(context.Background(), testItem())

	require.NoError(t, err)
}

func TestCreateComputesOutputFromNameAndValue(t *testing.T) {
	log := mocks.NewMockOperationLog(t)
	expectAppend(log, "create")

	result, err := providers.NewItemProvider(log).Create(context.Background(), testItem())

	require.NoError(t, err)
	require.Equal(t, "first-one", result.Output)
}

func TestCreateFailsWhenTheRecordCannotBeAppended(t *testing.T) {
	appendErr := errors.New("disk full")

	log := mocks.NewMockOperationLog(t)
	log.EXPECT().
		Append(mock.Anything, "ops.jsonl", oplog.Operation{Operation: "create", Address: "test.item.first"}).
		Return(appendErr).
		Once()

	item := testItem()

	result, err := providers.NewItemProvider(log).Create(context.Background(), item)

	require.ErrorIs(t, err, appendErr)
	require.Nil(t, result)
	require.Empty(t, item.Output)
}

func TestReadRecordsTheRead(t *testing.T) {
	log := mocks.NewMockOperationLog(t)
	expectAppend(log, "read")

	_, err := providers.NewItemProvider(log).Read(context.Background(), testItem(), testItem())

	require.NoError(t, err)
}

func TestReadKeepsTheOutputItWasCreatedWith(t *testing.T) {
	log := mocks.NewMockOperationLog(t)
	expectAppend(log, "read")

	old := testItem()
	old.Output = "first-one"

	updated := testItem()
	updated.Value = "changed"

	result, err := providers.NewItemProvider(log).Read(context.Background(), old, updated)

	require.NoError(t, err)
	require.Equal(t, "first-one", result.Output)
	require.Equal(t, "changed", result.Value)
}

func TestReadFailsWhenTheRecordCannotBeAppended(t *testing.T) {
	appendErr := errors.New("disk full")

	log := mocks.NewMockOperationLog(t)
	log.EXPECT().
		Append(mock.Anything, "ops.jsonl", oplog.Operation{Operation: "read", Address: "test.item.first"}).
		Return(appendErr).
		Once()

	result, err := providers.NewItemProvider(log).Read(context.Background(), testItem(), testItem())

	require.ErrorIs(t, err, appendErr)
	require.Nil(t, result)
}

func TestChangedAnswersUpdateWhenValueChanges(t *testing.T) {
	provider := providers.NewItemProvider(mocks.NewMockOperationLog(t))

	old := testItem()
	updated := testItem()
	updated.Value = "two"

	changes := []entity.PropertyChange{
		{Path: entity.Path{}.Attribute("value"), Before: "one", After: "two"},
	}

	change, err := provider.Changed(context.Background(), old, updated, changes, nil)

	require.NoError(t, err)
	require.Equal(t, entity.Update, change)
}

func TestChangedAnswersUpdateWhenSecretChanges(t *testing.T) {
	provider := providers.NewItemProvider(mocks.NewMockOperationLog(t))

	old := testItem()
	old.Secret = types.NewSensitive("hunter2")

	updated := testItem()
	updated.Secret = types.NewSensitive("hunter3")

	changes := []entity.PropertyChange{
		{Path: entity.Path{}.Attribute("secret"), Before: "hunter2", After: "hunter3", Sensitive: true},
	}

	change, err := provider.Changed(context.Background(), old, updated, changes, nil)

	require.NoError(t, err)
	require.Equal(t, entity.Update, change)
}

func TestChangedAnswersNoChangeWhenNothingChanged(t *testing.T) {
	provider := providers.NewItemProvider(mocks.NewMockOperationLog(t))

	change, err := provider.Changed(context.Background(), testItem(), testItem(), nil, nil)

	require.NoError(t, err)
	require.Equal(t, entity.NoChange, change)
}

func TestChangedNeverReplacesForAReplacedDependency(t *testing.T) {
	provider := providers.NewItemProvider(mocks.NewMockOperationLog(t))

	dependencies := []entity.DependencyChange{
		{Address: "test.item.other", Change: entity.Replace},
	}

	change, err := provider.Changed(context.Background(), testItem(), testItem(), nil, dependencies)

	require.NoError(t, err)
	require.NotEqual(t, entity.Replace, change)
}

func TestUpdateRecordsTheUpdate(t *testing.T) {
	log := mocks.NewMockOperationLog(t)
	expectAppend(log, "update")

	changes := []entity.PropertyChange{
		{Path: entity.Path{}.Attribute("value"), Before: "one", After: "two"},
	}

	_, err := providers.NewItemProvider(log).Update(context.Background(), testItem(), changes, nil)

	require.NoError(t, err)
}

func TestUpdateRecomputesOutputFromTheNewValue(t *testing.T) {
	log := mocks.NewMockOperationLog(t)
	expectAppend(log, "update")

	item := testItem()
	item.Output = "first-one"
	item.Value = "two"

	changes := []entity.PropertyChange{
		{Path: entity.Path{}.Attribute("value"), Before: "one", After: "two"},
	}

	result, err := providers.NewItemProvider(log).Update(context.Background(), item, changes, nil)

	require.NoError(t, err)
	require.Equal(t, "first-two", result.Output)
}

func TestUpdateFailsWhenTheRecordCannotBeAppended(t *testing.T) {
	appendErr := errors.New("disk full")

	log := mocks.NewMockOperationLog(t)
	log.EXPECT().
		Append(mock.Anything, "ops.jsonl", oplog.Operation{Operation: "update", Address: "test.item.first"}).
		Return(appendErr).
		Once()

	item := testItem()
	item.Output = "first-one"
	item.Value = "two"

	changes := []entity.PropertyChange{
		{Path: entity.Path{}.Attribute("value"), Before: "one", After: "two"},
	}

	result, err := providers.NewItemProvider(log).Update(context.Background(), item, changes, nil)

	require.ErrorIs(t, err, appendErr)
	require.Nil(t, result)
	require.Equal(t, "first-one", item.Output)
}

func TestDestroyRecordsTheDestroy(t *testing.T) {
	log := mocks.NewMockOperationLog(t)
	expectAppend(log, "destroy")

	err := providers.NewItemProvider(log).Destroy(context.Background(), testItem(), false)

	require.NoError(t, err)
}

func TestDestroyFailsWhenTheRecordCannotBeAppended(t *testing.T) {
	appendErr := errors.New("disk full")

	log := mocks.NewMockOperationLog(t)
	log.EXPECT().
		Append(mock.Anything, "ops.jsonl", oplog.Operation{Operation: "destroy", Address: "test.item.first"}).
		Return(appendErr).
		Once()

	err := providers.NewItemProvider(log).Destroy(context.Background(), testItem(), false)

	require.ErrorIs(t, err, appendErr)
}

func TestFunctionsOffersNone(t *testing.T) {
	provider := providers.NewItemProvider(mocks.NewMockOperationLog(t))

	require.Nil(t, provider.Functions())
}
