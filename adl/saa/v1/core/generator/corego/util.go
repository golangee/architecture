package corego

import (
	"github.com/golangee/src/ast"
	"strings"
)

// ModName returns the modules name.
func ModName(n ast.Node) string {
	var mod *ast.Mod
	if ast.ParentAs(n, &mod) {
		return mod.Name
	}

	return ""
}

// PkgName returns the parents package name.
func PkgName(n ast.Node) string {
	if p, ok := n.(*ast.Pkg); ok {
		elems := strings.Split(p.Path, "/")
		if len(elems) == 0 {
			return ""
		}

		return strings.Join(elems[:len(elems)-1], "/")
	}

	var pkg *ast.Pkg
	if ast.ParentAs(n, &pkg) {
		return PkgName(pkg)
	}

	return ""
}

// PkgRelativeName returns the relative path within the given module name.
//
func PkgRelativeName(n ast.Node) string {
	modName := ModName(n)
	pkgName := PkgName(n)

	return pkgName[len(modName)+1:]
}

func ShortModName(n ast.Node) string {
	name := ModName(n)
	elems := strings.Split(name, "/")
	if len(elems) > 0 {
		return elems[len(elems)-1]
	}

	return name
}
