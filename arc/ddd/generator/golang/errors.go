package golang

import (
	"github.com/golangee/architecture/arc/adl"
	"github.com/golangee/architecture/arc/generator/astutil"
	"github.com/golangee/src/ast"
	"github.com/golangee/src/golang"
	"github.com/golangee/src/stdlib/lang"
)

func buildErrors(parent *ast.File, srcMod *adl.Module, srcBc *adl.BoundedContext, src *adl.Package, errors []*adl.Error) (*lang.Error, error) {
	pkgSumType := lang.NewError(golang.MakePublic(srcBc.Name.String()))
	for _, anErr := range errors {
		errCase := lang.NewErrorCase(anErr.Name.String()).SetComment(anErr.Comment.String())
		for _, field := range anErr.Fields {
			errCase.AddProperty(field.Name.String(), astutil.MakeTypeDecl(field.Type), field.Comment.String())
		}

		pkgSumType.AddCase(errCase)
	}

	parent.AddNodes(pkgSumType.TypeDecl())

	return pkgSumType, nil
}
