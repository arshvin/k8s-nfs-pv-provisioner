package storage

import "testing"

func checkStringTestResults(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("Expected value: %v but actual: %v", expected, actual)
	}
}

func checkBoolTestResults(t *testing.T, description string, expected, actual bool) {
	if expected != actual {
		t.Errorf("Input string for test: '%v', Expected value: %v but actual: %v", description, expected, actual)
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

func Test_checkMatchTrueStr(t *testing.T) {
	checkBoolTestResults(t, "true", true, checkMatchTrueStr("true"))
	checkBoolTestResults(t, "TRUE", true, checkMatchTrueStr("TRUE"))
	checkBoolTestResults(t, "yes", true, checkMatchTrueStr("yes"))
	checkBoolTestResults(t, "YES", true, checkMatchTrueStr("YES"))

	checkBoolTestResults(t, "nope", false, checkMatchTrueStr("nope"))
	checkBoolTestResults(t, "false", false, checkMatchTrueStr("false"))
	checkBoolTestResults(t, "true|true", false, checkMatchTrueStr("true|true"))
	checkBoolTestResults(t, "empty string", false, checkMatchTrueStr(""))
}

func Test_checkMatchDNSorIPV4(t *testing.T){
	checkBoolTestResults(t, "10.0.0.1", true, checkMatchDNSorIPV4("10.0.0.1"))
	checkBoolTestResults(t, "domain-1.domain-2.domain-3", true, checkMatchDNSorIPV4("domain-1.domain-2.domain-3"))

	checkBoolTestResults(t, "domain-1.domain-2.domain_3", false, checkMatchDNSorIPV4("domain-1.domain-2.domain_3"))
	checkBoolTestResults(t, "/domain-1.domain-2.domain", false, checkMatchDNSorIPV4("/domain-1.domain-2.domain"))
}
