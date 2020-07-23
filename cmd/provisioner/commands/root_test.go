package commands

import (
	"testing"

	storage_v1 "k8s.io/api/storage/v1"
)

func TestSelectingStorageClass(t *testing.T) {
	clusterClasses := make([]storage_v1.StorageClass, 3)
	clusterClasses[0] = storage_v1.StorageClass{}
	clusterClasses[0].Name = "sc1"

	clusterClasses[1] = storage_v1.StorageClass{}
	clusterClasses[1].Name = "sc2"

	clusterClasses[2] = storage_v1.StorageClass{}
	clusterClasses[2].Name = "sc3"

	cliClasses := make([]string, 3)
	cliClasses[0] = "sc2"
	cliClasses[1] = "sc1"

	expected := make([]storage_v1.StorageClass, 2)
	expected[0] = clusterClasses[0]
	expected[1] = clusterClasses[1]

	actual,_ := selectClasses(clusterClasses, cliClasses)

	if len(expected) != len(actual) {
		t.Fatalf("Length of Expected slice:'%v' but lenght of Actual slice:'%v'", len(expected), len(actual))
	}
	if expected[0].Name != actual[0].Name {
		t.Fatalf("Expected[0] value:'%v' Actual[0] value:'%v'", expected[0].Name, actual[0].Name)
	}
	if expected[1].Name != actual[1].Name {
		t.Fatalf("Expected[1] value:'%v' Actual[1] value:'%v'", expected[1].Name, actual[1].Name)
	}

	cliClasses = append(cliClasses, "sc_absent")
	_, err := selectClasses(clusterClasses, cliClasses)
	if err == nil {
		t.Fatalf("The item '%v' is absent and the test must be failed", cliClasses[2])
	}

}
