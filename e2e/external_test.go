package e2e

import "testing"

func TestExternalAppliesPlansNoChangesAndDestroys(t *testing.T) {
	appliesPlansNoChangesAndDestroys(t, newExternalConfig)
}

func TestExternalRecordsCreatesAndDestroysInDependencyOrder(t *testing.T) {
	recordsCreatesAndDestroysInDependencyOrder(t, newExternalConfig)
}

func TestExternalComputesOutputOnCreate(t *testing.T) {
	computesOutputOnCreate(t, newExternalConfig)
}

func TestExternalUpdatesValueInPlaceAndRecomputesOutput(t *testing.T) {
	updatesValueInPlaceAndRecomputesOutput(t, newExternalConfig)
}

func TestExternalTreatsTheSecretAsSensitive(t *testing.T) {
	treatsTheSecretAsSensitive(t, newExternalConfig)
}
