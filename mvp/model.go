package mvp

import (
	"encoding/json"
	"fmt"
	"github.com/golangee/dyml"
	"github.com/golangee/dyml/parser"
	"github.com/golangee/dyml/token"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const DomainMetaFile = "meta.dyml"

// PString is a string with positional information.
type PString struct {
	Value string
	Range token.Position
}

func (p *PString) UnmarshalDyml(node *parser.TreeNode) error {
	if len(node.Children) == 0 {
		return token.NewPosError(node.Range, "expected a string")
	}

	text := node.Children[0]

	if !text.IsText() {
		return token.NewPosError(node.Range, "expected a string")
	}

	p.Value = *text.Text
	p.Range = text.Range

	return nil
}

// Domain is the tree root for all parsed information.
type Domain struct {
	Name        PString    `dyml:"name"`
	ArcVersion  ArcVersion `dyml:"arc_version"`
	Description PString    `dyml:"description"`
	// Executables is a list of executables. There must be exactly one in this prototype.
	Executables     []Executable `dyml:"executable"`
	BoundedContexts []BoundedContext
}

type Executable struct {
	Name         PString
	Architecture ArchitecturalModel `dyml:"architecture"`
	Generators   GeneratorSelection `dyml:"generator"`
	License      License            `dyml:"license"`
}

// UnmarshalDyml for an Executable will read a node as its name and them unmarshal all children into
// the Executable itself.
func (e *Executable) UnmarshalDyml(node *parser.TreeNode) error {
	if len(node.Children) != 1 {
		return token.NewPosError(node.Range, "define only a single executable here")
	}

	exe := node.Children[0]
	if !exe.IsNode() {
		return token.NewPosError(node.Range, "expected a node as an executable definition")
	}

	e.Name = PString{
		Value: exe.Name,
		Range: exe.Range,
	}

	// Try to unmarshal children into the remaining fields.
	// Ignore parsing errors here, validation later should catch missing values.
	for _, child := range exe.Children {
		_ = dyml.UnmarshalTree(child, &e.Architecture, false)
		_ = dyml.UnmarshalTree(child, &e.Generators, false)
		_ = dyml.UnmarshalTree(child, &e.License, false)
	}

	return nil
}

type GeneratorSelection struct {
	Go *GoGeneratorOptions `dyml:"go"`
}

type GoGeneratorOptions struct {
	// Version is the version of go to use for this project.
	Version      PString        `dyml:"version"`
	Dependencies []GoDependency `dyml:"dependency"`
}

type GoDependency struct {
	Name    PString `dyml:"name"`
	Version PString `dyml:"version"`
}

type ArchitecturalModel struct {
	Type PString `dyml:"type"`
}

type License struct {
	Name PString `dyml:",inner"`
	// downloadError is an error that may be set when an error occurred during download.
	downloadError error
	downloadText  string
}

type ArcVersion struct {
	V PString `dyml:",inner"`
}

// BoundedContext has no packages in this prototype.
type BoundedContext struct {
	Name     string
	Glossary []Glossary
	Stories  []Story      `dyml:"story"`
	Services []Service    `dyml:"service"`
	DTOs     []DTO        `dyml:"dto"`
	Imports  []TypeImport `dyml:"import"`
}

// ResolveType returns a fully resolved qualification for the given identifier.
// Might return nil if the type could not be resolved.
func (b *BoundedContext) ResolveType(domain *Domain, name string) TypeResolver {
	firstDotIdx := strings.Index(name, ".")
	if firstDotIdx >= 0 {
		// The name contains dots, which means it's fully qualified.
		// Try to resolve the name in either this bounded context first or recursively resolve
		// from another one.
		typeBC := name[:firstDotIdx]
		if b.Name == typeBC {
			// The name should be in this bc, try to find it by splitting of the identifier.
			split := strings.LastIndex(name, ".")
			return b.ResolveType(domain, name[split:])
		} else {
			// The name must be in another BC
			for _, bc := range domain.BoundedContexts {
				if &bc == b {
					// Do not check self
					continue
				}

				resolved := bc.ResolveType(domain, name)
				if resolved != nil {
					return resolved
				}
			}
		}
	} else {
		// The name is not fully qualified, it must be in this BC.
		// It will either be a type defined in the BC or an imported type.
		for _, dto := range b.DTOs {
			if name == strings.TrimSpace(dto.Name.Value) {
				return &dto
			}
		}
		for _, imp := range b.Imports {
			if imp.Name == name {
				return &imp
			}
		}
	}

	return nil
}

// TypeResolver is an interface implemented by all identifier that must be resolved to a type.
type TypeResolver interface {
	// isResolveType is a sealed method
	isResolveType()
}

// TypeImport is a type that was imported from other packages.
type TypeImport struct {
	Name string        `dyml:"name,attr"`
	Go   *TypeImportGo `dyml:"go"`
}

func (t *TypeImport) isResolveType() {}

type TypeImportGo struct {
	Import PString `dyml:"import"`
	Type   PString `dyml:"type"`
}

type DTO struct {
	Name       PString
	Parameters []Parameter
}

func (d *DTO) isResolveType() {}

func (d *DTO) UnmarshalDyml(node *parser.TreeNode) error {
	if len(node.Children) != 1 {
		return token.NewPosError(node.Range, "expected an identifier as DTO name")
	}

	nameNode := node.Children[0]
	d.Name = PString{
		Value: nameNode.Name,
		Range: nameNode.Range,
	}

	// Look for different nodes
	var propertiesNode *parser.TreeNode

	for _, child := range nameNode.Children {
		if child.Name == "properties" {
			if propertiesNode == nil {
				propertiesNode = child
			} else {
				return token.NewPosError(child.Range, "properties defined multiple times")
			}
		}
	}

	if propertiesNode != nil {
		// Collect parameters from children, first level is name
		for _, paramName := range propertiesNode.Children {
			param := Parameter{
				Name: PString{
					Value: paramName.Name,
					Range: paramName.Range,
				},
			}

			var paramType *parser.TreeNode
			for _, child := range paramName.Children {
				if child.IsText() {
					param.Description += *child.Text
				} else if child.IsNode() {
					paramType = child
					break
				}
			}

			if paramType == nil {
				return token.NewPosError(paramName.Range, "exactly one type name expected")
			}

			param.Type = PString{
				Value: paramType.Name,
				Range: paramType.Range,
			}

			d.Parameters = append(d.Parameters, param)
		}
	}

	return nil
}

type Parameter struct {
	Name        PString
	Type        PString
	Description string
}

// Service is any service that provides methods.
type Service struct {
	Name string      `dyml:"name,attr"`
	Type ServiceType `dyml:"type"`
}

type ServiceType struct {
	V PString `dyml:",inner"`
}

type Glossary struct {
	Entries map[PString]PString
}

type Story struct {
	Name     string            `dyml:"name,attr"`
	Title    PString           `dyml:"title"`
	AsA      PString           `dyml:"as_a"`
	IWantTo  PString           `dyml:"i_want_to"`
	SoThat   PString           `dyml:"so_that"`
	Criteria []AcceptCriterion `dyml:"accept"`
}

type AcceptCriterion struct {
	Require []PString `dyml:"require"`
	When    []PString `dyml:"when"`
	Then    []PString `dyml:"then"`
}

func LoadDomain(folder string) (*Domain, error) {
	// Load meta information
	metaFile, err := os.Open(filepath.Join(folder, DomainMetaFile))
	if err != nil {
		return nil, err
	}
	defer metaFile.Close()

	domain := &Domain{}
	err = dyml.Unmarshal(metaFile, &domain, false)
	if err != nil {
		return nil, err
	}

	// Load other stuff that is in the domain's root folder.

	// We can ignore the error here, as the folder must exist because we just loaded the domain meta.
	domainFolderItems, _ := ioutil.ReadDir(folder)

	for _, item := range domainFolderItems {
		if item.IsDir() || item.Name() == DomainMetaFile {
			continue
		}

		file, err := os.Open(filepath.Join(folder, item.Name()))
		if err != nil {
			return nil, err
		}
		defer file.Close()

		err = dyml.Unmarshal(file, &domain, false)
		if err != nil {
			return nil, err
		}
	}

	// Load bounded contexts
	for _, item := range domainFolderItems {
		if !item.IsDir() {
			continue
		}

		bc, err := LoadBoundedContext(folder, item.Name())
		if err != nil {
			return nil, err
		}

		domain.BoundedContexts = append(domain.BoundedContexts, *bc)
	}

	return domain, nil
}

// LoadBoundedContext parses a bounded context from disk. It lives inside the folder
// bcName inside the domainRoot folder.
func LoadBoundedContext(domainRoot, bcName string) (*BoundedContext, error) {
	bc := &BoundedContext{Name: bcName}

	bcPath := filepath.Join(domainRoot, bcName)

	// Ignore the error, we know that the BC exists, otherwise we would not know its name.
	bcItems, _ := ioutil.ReadDir(bcPath)

	for _, item := range bcItems {
		if item.IsDir() {
			// Folder would be packages we would need to load.
			continue
		}

		itemPath := filepath.Join(bcPath, item.Name())
		file, err := os.Open(itemPath)
		if err != nil {
			return nil, err
		}

		err = dyml.Unmarshal(file, bc, false)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal file '%s' in BC '%s': %w", itemPath, bcName, err)
		}
	}

	return bc, nil
}

// GetText downloads the text for this license.
// This will only download once and cache the error and text for later calls.
func (l *License) GetText() (string, error) {
	if l.downloadError == nil && len(l.downloadText) == 0 {
		licenseName := strings.TrimSpace(l.Name.Value)
		licenseUrl := fmt.Sprintf("https://spdx.org/licenses/%s.json", licenseName)
		resp, err := http.Get(licenseUrl)
		if err != nil {
			l.downloadError = token.NewPosError(l.Name.Range, "unable to download license: "+err.Error())
		} else {
			var responseJson map[string]interface{}
			err := json.NewDecoder(resp.Body).Decode(&responseJson)
			if err != nil {
				l.downloadError = token.NewPosError(l.Name.Range, "invalid license name")
			} else {
				l.downloadText = responseJson["licenseText"].(string)
			}
		}
	}

	return l.downloadText, l.downloadError
}
