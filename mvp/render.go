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

const MakefileTemplate = `build: tidy
	go build cmd/{{.Name.Value}}.go
run: build
	./{{.Name.Value}}
tidy:
	go mod tidy
`

const ReadmeTemplate = `= {{.Executable.Name.Value}}
_Generated with prototype version {{trim .Domain.ArcVersion.V.Value}} of https://github.com/golangee/architecture_

All bounded contexts will be summarized here.

{{range .Domain.BoundedContexts}}
== {{.Name}}
Types: {{range .DTOs}}{{.Name.Value}}, {{else}} No types defined. {{end}}

=== User Stories
{{range .Stories}}
* {{.Name}}: {{trim .Title.Value}}
{{else}}
No user stories.
{{end}}
{{end}}
`

var templateUtilFunctions template.FuncMap

// Render creates a project that's ready to run for this domain into the given folder.
// No validation is run here, so you might want to call Validate on the domain before
// running this.
func (e *Executable) Render(domain *Domain, folder string) error {
	templateUtilFunctions = template.FuncMap{
		"trim": strings.TrimSpace,
	}

	folderAbs, err := filepath.Abs(folder)
	if err != nil {
		return err
	}

	err = os.MkdirAll(folderAbs, DefaultFileMode)
	if err != nil {
		return err
	}

	err = e.renderModFile(folderAbs)
	if err != nil {
		return err
	}

	err = e.renderMakefile(folderAbs)
	if err != nil {
		return err
	}

	err = e.renderLicense(folderAbs)
	if err != nil {
		return err
	}

	err = e.renderReadme(folderAbs, domain)
	if err != nil {
		return err
	}

	for _, bc := range domain.BoundedContexts {
		bcFolder := filepath.Join(folderAbs, "internal", bc.Name)

		err = os.MkdirAll(bcFolder, DefaultFileMode)
		if err != nil {
			return err
		}

		err = e.renderBoundedContext(bcFolder, &bc, domain)
		if err != nil {
			return err
		}
	}

	err = e.renderMain(folder, domain)
	if err != nil {
		return err
	}

	return nil
}

// renderMain generates an entrypoint for this application in the folder "<folder>/cmd/<execname>"
func (e *Executable) renderMain(folder string, domain *Domain) error {
	applicationFolder := filepath.Join(folder, "cmd")

	err := os.Mkdir(applicationFolder, DefaultFileMode)
	if err != nil {
		return err
	}

	main := jen.NewFile("main")
	main.Func().Id("main").Params().Block(
		jen.Qual("fmt", "Println").Call(jen.Lit("Hello generator!")),
	)

	mainFile, err := os.Create(filepath.Join(applicationFolder, fmt.Sprintf("%s.go", e.Name.Value)))
	if err != nil {
		return err
	}
	defer mainFile.Close()

	err = main.Render(mainFile)
	if err != nil {
		return err
	}

	return nil
}

// renderBoundedContext renders the given BoundedContext into the folder.
func (e *Executable) renderBoundedContext(folder string, bc *BoundedContext, domain *Domain) error {
	err := e.renderTypes(folder, bc, domain)
	if err != nil {
		return err
	}

	err = e.renderServices(folder, bc, domain)
	if err != nil {
		return err
	}

	return nil
}

// renderServices renders all services for a BoundedContext.
func (e *Executable) renderServices(folder string, bc *BoundedContext, domain *Domain) error {
	for _, serviceDefinition := range bc.Services {
		serviceName := strings.TrimSpace(serviceDefinition.Name)
		serviceFileName := filepath.Join(folder, fmt.Sprintf("%s.go", strcase.ToSnake(serviceName)))

		serviceFile, err := os.Create(serviceFileName)
		if err != nil {
			return err
		}

		service := jen.NewFile(bc.Name)
		service.Type().Id(serviceName).Struct()

		err = service.Render(serviceFile)
		if err != nil {
			return err
		}
	}

	return nil
}

// renderTypes renders all types for a BoundedContext into a single file.
// That's okay because types should (probably) never be edited directly.
func (e *Executable) renderTypes(folder string, bc *BoundedContext, domain *Domain) error {
	types := jen.NewFile(bc.Name)

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

	typesFile, err := os.Create(filepath.Join(folder, "types.gen.go"))
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

	readmeTemplate, err := template.New("readme").Funcs(templateUtilFunctions).Parse(ReadmeTemplate)
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

	temp, err := template.New("Makefile").Funcs(templateUtilFunctions).Parse(MakefileTemplate)
	if err != nil {
		return fmt.Errorf("invalid Makefile template: %w", err)
	}

	err = temp.Execute(makefile, e)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
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
