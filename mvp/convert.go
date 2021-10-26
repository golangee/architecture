package mvp

import (
	"fmt"
	"github.com/golangee/architecture/arc/adl"
	arctoken "github.com/golangee/architecture/arc/token"
	"github.com/iancoleman/strcase"
	"strings"
)

// Convert the given domain into a renderable adl.Project.
func Convert(domain *Domain) *adl.Project {
	return domain.convert()
}

func (d *Domain) convert() *adl.Project {
	return adl.NewProject(d.Name.Trimmed(), d.Description.Trimmed()).
		AddModules(d.defaultModule())
}

func (d *Domain) defaultModule() *adl.Module {
	return adl.NewModule("std", "the default module").
		SetGenerator(d.convertGenerator()).
		AddBoundedContexts(d.convertBoundedContexts()...).
		AddExecutables(d.convertExecutables()...)
}

func (d *Domain) convertGenerator() *adl.Generator {
	modName := d.modulePath()
	return adl.NewGenerator().
		SetOutDir("../testdata/supportiety_render").
		SetGo(
			adl.NewGolang().
				SetModName(modName).
				AddDist("darwin", "amd64"))
}

func (d *Domain) convertExecutables() []*adl.Executable {
	var result []*adl.Executable

	// TODO Convert executables

	return result
}

func (d *Domain) convertBoundedContexts() []*adl.BoundedContext {
	var result []*adl.BoundedContext

	for _, bc := range d.BoundedContexts {
		adlBC := adl.NewBoundedContext(bc.Name, "$MOD/"+bc.Name)
		result = append(result, adlBC)

		core := adl.NewPackage("std", "the default core package")

		for _, dto := range bc.DTOs {
			core.AddStructs(dto.convert())
		}

		adlBC.AddCore(core)

		usecase := adl.NewPackage("std", "the default usecase package")

		for _, service := range bc.Services {
			usecase.AddServices(adl.NewService(service.Name, ""))
		}

		adlBC.AddUsecase(usecase)
	}

	return result
}

func (d *Domain) modulePath() string {
	return fmt.Sprintf("github.com/golangee/architecture/testdata/supportiety_render/%s", strings.ToLower(d.Name.Trimmed()))
}

func (d DTO) convert() *adl.Struct {
	strct := adl.NewDTO(d.Name.Trimmed(), "")
	for _, parameter := range d.Parameters {
		strct.AddFields(adl.NewField(
			strcase.ToCamel(parameter.Name.Trimmed()),
			parameter.Description,
			adl.NewTypeDecl(parameter.Type.Trimmed())))
	}
	return strct
}

func (p PString) convert() arctoken.String {
	return arctoken.String{
		Position: arctoken.Position{
			BeginPos: arctoken.Pos{
				File:   p.Range.BeginPos.File,
				Offset: p.Range.BeginPos.Offset,
				Line:   p.Range.BeginPos.Line,
				Col:    p.Range.BeginPos.Col,
			},
			EndPos: arctoken.Pos{
				File:   p.Range.EndPos.File,
				Offset: p.Range.EndPos.Offset,
				Line:   p.Range.EndPos.Line,
				Col:    p.Range.EndPos.Col,
			},
		},
		Val: p.Value,
	}
}
