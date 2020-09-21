package storage

import "testing"

func checkStringTestResults(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("Expected value: %v but actual: %v", expected, actual)
	}
}

func Test_ChooseBaseNameOfAsset(t *testing.T) {
	checkStringTestResults(t, "some-namespace-some-pvc-vol", ChooseBaseNameOfAsset("some-namespace", "some-pvc"))
	checkStringTestResults(t, "some-namespace-some-pvc-vol", ChooseBaseNameOfAsset("some", "namespace", "some", "pvc"))
	checkStringTestResults(t, "some-namespace-some-pvc-claim-vol", ChooseBaseNameOfAsset("some-namespace", "some-pvc-claim"))
}

func Test_CastToInt(t *testing.T) {
	expected := 1000
	actual, _ := castToInt("1000")
	if expected != actual {
		t.Errorf("Expected value: %v but actual: %v", expected, actual)
	}

	if _, err := castToInt("FF"); err == nil {
		t.Errorf("The ERR should be not nil")
	}

	if _, err := castToInt("10O"); err == nil {
		t.Error("The ERR should be not nil")
	}
}

