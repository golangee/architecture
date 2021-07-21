package model

import (
	"encoding/json"
	"os"
	"testing"
)

// TODO Make a proper test, this one just writes the parsed structure to a file so I can look at it more easily.
func TestRead(t *testing.T) {
	proj, err := ParseWorkspace("/Users/larsmartens/repos/golangee/architecture/testdata/supportiety")
	if err != nil {
		t.Fatal(err)
	}

	indented, err := json.MarshalIndent(proj, "", "    ")
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create("test_out.json")
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	_, err = f.Write(indented)
	if err != nil {
		t.Fatal(err)
	}
}
