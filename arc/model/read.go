package model

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/golangee/tadl"
)

// ParseWorkspace parses a Tadl workspace folder at the given path.
func ParseWorkspace(folder string) (*Project, error) {
	// Read metadata for whole project
	f, err := os.Open(filepath.Join(folder, "meta.tadl"))
	if err != nil {
		return nil, err
	}

	var project Project
	if err = tadl.Unmarshal(f, &project, false); err != nil {
		return nil, err
	}

	// Read glossary
	f, err = os.Open(filepath.Join(folder, "glossary.tadl"))
	if err != nil {
		return nil, err
	}

	var glossary Glossary
	if err = tadl.Unmarshal(f, &glossary, false); err != nil {
		return nil, err
	}

	project.Glossary = glossary

	// Parse stories
	f, err = os.Open(filepath.Join(folder, "story.tadl"))
	if err != nil {
		return nil, err
	}

	if err = tadl.Unmarshal(f, &project, false); err != nil {
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
	f, err := os.Open(filepath.Join(folder, "meta.tadl"))
	if err != nil {
		return nil, err
	}

	var context BoundedContext
	if err := tadl.Unmarshal(f, &context, false); err != nil {
		return nil, err
	}

	// Authors
	f, err = os.Open(filepath.Join(folder, "authors.tadl"))
	if err != nil {
		return nil, err
	}

	var authors Authors
	if err = tadl.Unmarshal(f, &authors, false); err != nil {
		return nil, err
	}

	context.Authors = authors

	// Artifacts
	f, err = os.Open(filepath.Join(folder, "artifact.tadl"))
	if err != nil {
		return nil, err
	}

	var artifacts Artifacts
	if err = tadl.Unmarshal(f, &artifacts, false); err != nil {
		return nil, err
	}

	// Aggregate methods may have an incorrectly parsed parameter called "ret", that should only
	// be the return parameter. We fix that here.
	for _, aggregate := range artifacts.Aggregates {
		for name, method := range aggregate.Methods {
			// These two checks guarantuee that no parameter called "ret" is removed wrongly.
			if s, ok := method.Params["ret"]; ok {
				if _, ok := method.Returns[s]; ok {
					delete(method.Params, "ret")
					aggregate.Methods[name] = method
				}
			}
		}
	}

	context.Artifacts = artifacts

	return &context, nil
}
