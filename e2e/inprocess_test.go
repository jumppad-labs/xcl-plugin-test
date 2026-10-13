package e2e

import "testing"

func TestInProcessAppliesPlansNoChangesAndDestroys(t *testing.T) {
	appliesPlansNoChangesAndDestroys(t, newInProcessConfig)
}

func TestInProcessRecordsCreatesAndDestroysInDependencyOrder(t *testing.T) {
	recordsCreatesAndDestroysInDependencyOrder(t, newInProcessConfig)
}

func TestInProcessComputesOutputOnCreate(t *testing.T) {
	computesOutputOnCreate(t, newInProcessConfig)
}

func TestInProcessUpdatesValueInPlaceAndRecomputesOutput(t *testing.T) {
	updatesValueInPlaceAndRecomputesOutput(t, newInProcessConfig)
}

func TestInProcessTreatsTheSecretAsSensitive(t *testing.T) {
	treatsTheSecretAsSensitive(t, newInProcessConfig)
}
