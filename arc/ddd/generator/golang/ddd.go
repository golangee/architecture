package golang

import (
	"fmt"
	"github.com/golangee/architecture/arc/adl"
	"github.com/golangee/architecture/arc/generator/astutil"
	"github.com/golangee/architecture/arc/generator/doc"
	"github.com/golangee/architecture/arc/generator/golang"
	"github.com/golangee/architecture/arc/generator/stereotype"
	"github.com/golangee/architecture/arc/token"
	"github.com/golangee/src/ast"
	golang2 "github.com/golangee/src/golang"
	"strings"
)

func RenderModule(dst *ast.Prj, prj *adl.Project, src *adl.Module) error {
	normalizeNames(src)

	if src.Generator == nil {
		return fmt.Errorf("cannot render a non-target project: %s", src.Name)
	}

	if src.Generator.Go == nil {
		return fmt.Errorf("cannot render a non-go module: %s -> %s", src.Name, src.Generator.OutDir)
	}

	mod := astutil.MkMod(dst, src.Generator.Go.Module.String()).
		SetLang(ast.LangGo).
		SetLangVersion(ast.LangVersionGo16).
		SetOutputDirectory(src.Generator.OutDir.String())

	mod.Require("github.com/golangee/log latest")

	stereotype.ModFrom(mod).SetProjectIdent(prj.Name.String())
	stereotype.ModFrom(mod).SetIdent(src.Name.String())

	// set the root doc (lets use the project context)
	stereotype.ModFrom(mod).Docs().Append("docs/content/_index.md",
		doc.NewElement("h2").Append(doc.NewText(prj.Name.String())),
		doc.NewText(golang2.DeEllipsis(prj.Name.String(), prj.Comment.String())),
	)

	// set the module doc
	stereotype.Doc(mod, "", "_index.md",
		doc.NewElement("h2").Append(doc.NewText(src.Name.String())),
		doc.NewText(golang2.DeEllipsis(prj.Name.String(), src.Comment.String())),
	)

	stereotype.ModFrom(mod).SetIdent(src.Name.String())
	stereotype.Doc(mod, "", "getting-started/_index.md", doc.NewText("everything you must need to know to get started."))

	for _, require := range src.Generator.Go.Requires {
		mod.Require(require.String())
	}

	// logger
	if err := renderLogger(mod, src); err != nil {
		return token.NewPosError(src.Name, "cannot render logger").SetCause(err)
	}

	// buildinfo
	if err := renderBuildInfo(mod, src); err != nil {
		return token.NewPosError(src.Name, "cannot render build info").SetCause(err)
	}

	// execs
	if err := renderExecs(mod, src); err != nil {
		return token.NewPosError(src.Name, "cannot render executable entry points").SetCause(err)
	}

	// makefile
	if err := renderMakefile(mod, src); err != nil {
		return token.NewPosError(src.Name, "cannot render makefile").SetCause(err)
	}

	// bounded context packages
	for _, bc := range src.BoundedContexts {
		domainTerm := prj.Glossary.Terms[bc.Name.String()]
		domainRootPkg := golang.MakePkgPath(bc.Path.String())
		domain := astutil.MkPkg(mod, domainRootPkg)
		domain.SetComment("...contains the core and usecase packages which represent the bounded contexts" + bc.Name.String() + " domain model.\n" + golang2.DeEllipsis(domainTerm.Ident.String(), domainTerm.Description.String()))
		domain.SetPreamble(makePreamble(src.Preamble))

		// core packages
		coreRootPkg := golang.MakePkgPath(domainRootPkg, "core")
		coreRoot := astutil.MkPkg(mod, coreRootPkg)
		coreRoot.SetComment("...contains the domains primitives like Entities, Values, Repositories, Services (anemic), Events, aggregate roots (non-anemic) or DTOs.")
		coreRoot.SetPreamble(makePreamble(src.Preamble))
		for _, p := range bc.Core {
			if err := renderUserTypes(coreRoot, src, bc, p); err != nil {
				return token.NewPosError(p.Name, "unable to render domain core package").SetCause(err)
			}
		}

		// usecase packages
		usecaseRootPkg := golang.MakePkgPath(domainRootPkg, "usecase")
		usecaseRoot := astutil.MkPkg(mod, usecaseRootPkg)
		usecaseRoot.SetComment("...contains the domains use cases, usually in a service form, which uses an arbitrary composition of the domains primitives.")
		usecaseRoot.SetPreamble(makePreamble(src.Preamble))
		for _, p := range bc.Usecase {
			if err := renderUserTypes(usecaseRoot, src, bc, p); err != nil {
				return token.NewPosError(p.Name, "unable to render domain usecase package").SetCause(err)
			}
		}
	}

	// build super configuration
	if err := renderConfigs(mod, src); err != nil {
		return token.NewPosError(src.Name, "cannot render configurations").SetCause(err)
	}

	// actual app and di layer
	if err := renderApps(mod, src); err != nil {
		return token.NewPosError(src.Name, "cannot render application").SetCause(err)
	}

	return nil
}

func renderUserTypes(parent *ast.Pkg, srcMod *adl.Module, srcBc *adl.BoundedContext, src *adl.Package) error {
	mod := astutil.Mod(parent)
	pkg := parent
	if src.Name.String() != "" {
		pkg = astutil.MkPkg(mod, golang.MakePkgPath(pkg.Path, src.Name.String()))
		pkg.SetPreamble(makePreamble(srcMod.Preamble))
		pkg.SetComment(src.Comment.String())
	}

	// errors
	if len(src.Errors) > 0 {
		file := ast.NewFile(strings.ToLower("errors.go"))
		pkg.AddFiles(file)
		file.SetPreamble(makePreamble(srcMod.Preamble))

		if _, err := buildErrors(file, srcMod, srcBc, src, src.Errors); err != nil {
			return fmt.Errorf("cannot build errors: %w", err)
		}

	}

	// dtos
	if len(src.DTOs) > 0 {
		file := ast.NewFile(strings.ToLower("dtos.go"))
		pkg.AddFiles(file)
		file.SetPreamble(makePreamble(srcMod.Preamble))
		for _, dto := range src.DTOs {
			typ, err := golang.AddComponent(file, dto)
			if err != nil {
				return err
			}

			_ = typ
		}
	}

	// repos
	if len(src.Repositories) > 0 {
		file := ast.NewFile(strings.ToLower("repositories.go"))
		file.SetPreamble(makePreamble(srcMod.Preamble))
		pkg.AddFiles(file)

		for _, repository := range src.Repositories {
			iface, err := buildInterface(file, srcMod, src, repository)
			if err != nil {
				return err
			}

			for _, d := range repository.CRUDs {
				_, err := renderCrud(file, iface, d)
				if err != nil {
					return fmt.Errorf("unable to render CRUD: %w", err)
				}
			}
		}
	}

	// services
	if len(src.Services) > 0 {
		file := ast.NewFile(strings.ToLower("services.go"))
		pkg.AddFiles(file)
		file.SetPreamble(makePreamble(srcMod.Preamble))
		for _, srv := range src.Services {
			t, err := golang.AddComponent(file, srv.Component)
			stereotype.StructFrom(t).SetIsService(true)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func buildInterface(parent *ast.File, srcMod *adl.Module, src *adl.Package, iface *adl.Interface) (*ast.Interface, error) {
	aType := ast.NewInterface(iface.Name.String()).SetComment(iface.Comment.String())
	parent.AddTypes(aType)

	for _, method := range iface.Methods {
		aMethod := ast.NewFunc(method.Name.String()).SetComment(method.Comment.String())
		for _, param := range method.In {
			aMethod.AddParams(ast.NewParam(param.Name.String(), astutil.MakeTypeDecl(param.Type)).SetComment(param.Comment.String()))
		}

		for _, param := range method.Out {
			aMethod.AddResults(ast.NewParam(param.Name.String(), astutil.MakeTypeDecl(param.Type)).SetComment(param.Comment.String()))
		}

		aType.AddMethods(aMethod)
	}

	return aType, nil
}
