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
		t.Fatal("parser error:", err)
	}

	if err := proj.Validate(proj); err != nil {
		t.Fatal("validation failed:", err)
	}

	indented, err := json.MarshalIndent(proj, "", "    ")
	if err != nil {
		t.Fatal("could not create marshalled json:", err)
	}

	f, err := os.Create("test_out.json")
	if err != nil {
		t.Fatal("failed to create output file:", err)
	}

	defer f.Close()

	_, err = f.Write(indented)
	if err != nil {
		t.Fatal("failed to write to file", err)
	}
}
