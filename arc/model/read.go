package model

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/golangee/dyml"
)

// ParseWorkspace parses a Dyml workspace folder at the given path.
func ParseWorkspace(folder string) (*Project, error) {
	// Read metadata for whole project
	f, err := os.Open(filepath.Join(folder, "meta.dyml"))
	if err != nil {
		return nil, err
	}

	var project Project
	if err = dyml.Unmarshal(f, &project, false); err != nil {
		return nil, err
	}

	// Read glossary
	f, err = os.Open(filepath.Join(folder, "glossary.dyml"))
	if err != nil {
		return nil, err
	}

	var glossary Glossary
	if err = dyml.Unmarshal(f, &glossary, false); err != nil {
		return nil, err
	}

	project.Glossary = glossary

	// Read executables
	f, err = os.Open(filepath.Join(folder, "executable.dyml"))
	if err != nil {
		return nil, err
	}

	// TODO Parse stories

	if err = dyml.Unmarshal(f, &project, false); err != nil {
		return nil, err
	}

	// Iterate over folders as bounded contexts.
	folderItems, err := ioutil.ReadDir(folder)
	if err != nil {
		return nil, err
	}

	for _, item := range folderItems {
		if item.IsDir() {
			boundedContext, err := ParseBoundedContext(filepath.Join(folder, item.Name()))
			if err != nil {
				return nil, err
			}

			project.BoundedContexts = append(project.BoundedContexts, *boundedContext)
		}
	}

	return &project, nil
}

func ParseBoundedContext(folder string) (*BoundedContext, error) {
	// Parse meta information
	f, err := os.Open(filepath.Join(folder, "meta.dyml"))
	if err != nil {
		return nil, err
	}

	var context BoundedContext
	if err := dyml.Unmarshal(f, &context, false); err != nil {
		return nil, err
	}

	return &context, nil
}
