package mvp

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidateDomain(t *testing.T) {
	inputFolder, err := filepath.Abs("../testdata/supportiety")
	if err != nil {
		t.Fatal(err)
	}

	outputFolder, err := filepath.Abs("../testdata/supportiety_render")
	if err != nil {
		t.Fatal(err)
	}

	// Delete old outputFolder before doing anything
	err = os.RemoveAll(outputFolder)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("🗑 Deleted old render directory")

	domain, err := LoadDomain(inputFolder)
	if err != nil {
		t.Fatalf("%+v\n", err)
	} else {
		// Dump JSON to make it easier to look at.
		file, err := os.Create("model_test_out.json")
		if err != nil {
			log.Fatalln(err)
		}
		defer file.Close()

		err = json.NewEncoder(file).Encode(domain)
		if err != nil {
			t.Fatal(err)
		}
	}
	log.Println("📚 LoadDomain ok")

	err = domain.Validate(domain)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("🔍 Validate ok")

	err = domain.Executables[0].Render(domain, outputFolder)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("✏️ Render ok")
}
