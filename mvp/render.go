package mvp

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"github.com/golangee/dyml/token"
	"github.com/iancoleman/strcase"
	"golang.org/x/mod/modfile"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const DefaultFileMode = 0755

const MainDocFileTemplate = `%s

(Generated with prototype version %v of github.com/golangee/architecture)

%s
`

const MakefileTemplate = `run: tidy
	go run main.go
tidy:
	go mod tidy
`

const ReadmeTemplate = `= {{.Executable.Name.Value}}

All bounded contexts will be summarized here.

{{range .Domain.BoundedContexts}}
== {{.Name}}
Types: {{range .DTOs}}{{.Name.Value}}, {{else}} No types defined. {{end}}

=== User Stories
{{range .Stories}}
* {{.Name}}: {{.Title.Value -}}
{{else}}
No user stories for this bounded context.
{{end}}

{{end}}
`

// Render creates a project that's ready to run for this domain into the given folder.
// No validation is run here, so you might want to call Validate on the domain before
// running this.
func (e *Executable) Render(domain *Domain, folder string) error {
	err := os.MkdirAll(folder, DefaultFileMode)
	if err != nil {
		return err
	}

	err = e.renderMainDocFile(folder, domain)
	if err != nil {
		return err
	}

	err = e.renderModFile(folder)
	if err != nil {
		return err
	}

	err = e.renderMakefile(folder)
	if err != nil {
		return err
	}

	err = e.renderLicense(folder)
	if err != nil {
		return err
	}

	err = e.renderTypes(folder, domain)
	if err != nil {
		return err
	}

	err = e.renderReadme(folder, domain)
	if err != nil {
		return err
	}

	return nil
}

// renderTypes renders all types into a single file.
// That's okay because types should (probably) never be edited directly.
func (e *Executable) renderTypes(folder string, domain *Domain) error {
	// Emit file containing all types
	types := jen.NewFile("main")
	for _, bc := range domain.BoundedContexts {
		for _, dto := range bc.DTOs {
			var fields []jen.Code

			for _, parameter := range dto.Parameters {
				name := strings.TrimSpace(parameter.Name.Value)
				comment := strings.TrimSpace(parameter.Description)
				typ := strings.TrimSpace(parameter.Type.Value)

				resolved := bc.ResolveType(domain, typ)

				if len(comment) > 0 {
					fields = append(fields, jen.Comment(comment))
				}
				switch r := resolved.(type) {
				case *DTO:
					fields = append(fields, jen.Id(name).Id(typ))
				case *TypeImport:
					fields = append(fields, jen.Id(name).Qual(r.Go.Import.Value, r.Go.Type.Value))
				default:
					// Emit type name as is, hoping that it is a native type.
					// TODO Ideally a translation from abstract arc-types to language specific types would happen here.
					fields = append(fields, jen.Id(name).Id(typ))
				}
			}

			types.Type().Id(dto.Name.Value).Struct(fields...)
		}
	}

	typesFile, err := os.Create(filepath.Join(folder, "types.go"))
	if err != nil {
		return err
	}
	defer typesFile.Close()

	err = types.Render(typesFile)
	if err != nil {
		return err
	}

	return nil
}

// renderReadme creates a README.adoc with some rudimentary documentation.
func (e *Executable) renderReadme(folder string, domain *Domain) error {
	readme, err := os.Create(filepath.Join(folder, "README.adoc"))
	if err != nil {
		return fmt.Errorf("failed to create README: %w", err)
	}
	defer readme.Close()

	readmeTemplate, err := template.New("readme").Parse(ReadmeTemplate)
	if err != nil {
		return fmt.Errorf("bug: invalid template for README: %w", err)
	}

	err = readmeTemplate.Execute(readme, struct {
		Domain     *Domain
		Executable *Executable
	}{
		Domain:     domain,
		Executable: e,
	})
	if err != nil {
		return fmt.Errorf("failed to execute README template: %w", err)
	}

	return nil
}

func (e *Executable) renderLicense(folder string) error {
	licenseText, err := e.License.GetText()
	if err != nil {
		return err
	}

	licenseFile, err := os.Create(filepath.Join(folder, "LICENSE"))
	if err != nil {
		return err
	}
	defer licenseFile.Close()

	_, err = licenseFile.WriteString(licenseText)
	if err != nil {
		return err
	}

	return nil
}

func (e *Executable) renderMakefile(folder string) error {
	makefile, err := os.Create(filepath.Join(folder, "Makefile"))
	if err != nil {
		return fmt.Errorf("failed to create Makefile: %w", err)
	}

	_, err = makefile.WriteString(fmt.Sprintf(MakefileTemplate))
	if err != nil {
		return fmt.Errorf("failed to write Makefile: %w", err)
	}

	return nil
}

// renderModFiles renders the go.mod file
func (e *Executable) renderModFile(folder string) error {
	mod := modfile.File{
		Syntax: &modfile.FileSyntax{},
	}

	// First write the preamble with go version and module name
	err := mod.AddGoStmt(e.Generators.Go.Version.Value)
	if err != nil {
		return token.NewPosError(e.Generators.Go.Version.Range, fmt.Sprintf("invalid go version: %s", err.Error()))
	}

	err = mod.AddModuleStmt(strcase.ToLowerCamel(e.Name.Value))
	if err != nil {
		return token.NewPosError(e.Name.Range, fmt.Sprintf("name could not be formatted as a go module name: %s", err.Error()))
	}

	// Write all things to import
	for _, dependency := range e.Generators.Go.Dependencies {
		err := mod.AddRequire(strings.TrimSpace(dependency.Name.Value), strings.TrimSpace(dependency.Version.Value))
		if err != nil {
			return token.NewPosError(dependency.Name.Range, fmt.Sprintf("could not add dependency '%s': %s", dependency.Name.Value, err.Error()))
		}
	}

	bytes, err := mod.Format()
	if err != nil {
		return fmt.Errorf("failed to create go.mod file: %w", err)
	}

	modFile, err := os.Create(filepath.Join(folder, "go.mod"))
	if err != nil {
		return err
	}
	defer modFile.Close()

	_, err = modFile.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (e *Executable) renderMainDocFile(folder string, domain *Domain) error {
	doc := jen.NewFile("main")
	doc.PackageComment(fmt.Sprintf(MainDocFileTemplate, strings.TrimSpace(domain.Name.Value), strings.TrimSpace(domain.ArcVersion.V.Value), domain.Description.Value))

	docF, err := os.Create(filepath.Join(folder, "doc.go"))
	if err != nil {
		return err
	}
	defer docF.Close()

	err = doc.Render(docF)
	if err != nil {
		return err
	}

	return nil
}
