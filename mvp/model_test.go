package mvp

import (
	"encoding/json"
	"fmt"
	"github.com/golangee/architecture/arc"
	arctoken "github.com/golangee/architecture/arc/token"
	"github.com/golangee/src/render"
	"log"
	"os"
	"testing"
)

func TestRenderDomain(t *testing.T) {
	// Process input
	inputFolder := "../testdata/supportiety"

	outputFolder := "../testdata/supportiety_render"

	log.Println("📚 Loading domain...")
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
	log.Println("📚 ok")

	// Run validation
	log.Println("🔍 Validating domain...")
	err = domain.Validate(domain)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("🔍 ok")

	// Generate project
	log.Println("✏️ Rendering...")
	artifact, err := arc.Render(Convert(domain))

	if err != nil {
		fmt.Println(artifact)
		t.Fatal(arctoken.Explain(err))
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := render.Clean(outputFolder, []byte("Code generated by golangee/architecture; DO NOT EDIT.")); err != nil {
		t.Fatal(err)
	}

	if err := render.Write(cwd, artifact); err != nil {
		t.Fatal(err)
	}

	log.Println("✏️ ok")
}
