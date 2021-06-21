package golang

import (
	"github.com/golangee/architecture/arc/generator/astutil"
	"github.com/golangee/src/ast"
	"github.com/golangee/src/golang"
	"strings"
	"unicode"
)

// PkgPathDir is like filepath.Dir but always with /
func PkgPathDir(p string) string {
	segments := strings.Split(p, "/")
	if len(segments) == 0 {
		return p
	}

	return strings.Join(segments[:len(segments)-1], "/")
}

// PkgPathBase is like filepath.Base but always with /
func PkgPathBase(p string) string {
	segments := strings.Split(p, "/")
	if len(segments) == 0 {
		return p
	}

	return segments[len(segments)-1]
}

// MakePkgPath takes arbitrary fragments and creates a more or less idiomatic path of it.
func MakePkgPath(frags ...string) string {
	tmp := strings.Builder{}
	for i, f := range frags {
		subFrags := strings.Split(f, "/")
		for k, frag := range subFrags {
			if strings.HasPrefix(frag, "/") {
				frag = frag[1:]
			}

			if strings.HasSuffix(frag, "/") {
				frag = frag[:len(frag)-1]
			}

			frag = strings.ToLower(frag)
			frag = strings.ReplaceAll(frag, " ", "_")

			tmp.WriteString(strings.ToLower(frag))

			if k < len(subFrags)-1 {
				tmp.WriteString("/")
			}
		}

		if i < len(frags)-1 {
			tmp.WriteString("/")
		}

	}

	return tmp.String()
}

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

// GlobalFlatName tries to generate a readable and globally unique name without evaluating the actual context.
func GlobalFlatName(n ast.NamedType) string {
	const marker = "internal"
	fqn := astutil.FullQualifiedName(n)
	pos := strings.Index(fqn, marker)
	return MakePublic(golang.MakeIdentifier(fqn[pos+len(marker):]))
}

func MkFile(dst *ast.Prj, modName, pkgName, fname string) *ast.File {
	const preamble = "Code generated by golangee/architecture. DO NOT EDIT."

	mod := astutil.MkMod(dst, modName)
	mod.SetLang(ast.LangGo)
	pkg := astutil.MkPkg(mod, pkgName)
	file := astutil.MkFile(pkg, fname)
	file.SetPreamble(preamble)

	return file
}

// MakePublic converts aBc to ABc.
// Special cases:
//  * id becomes ID
func MakePublic(str string) string {
	if len(str) == 0 {
		return str
	}

	switch str {
	case "id":
		return "ID"
	default:
		return string(unicode.ToUpper(rune(str[0]))) + str[1:]
	}
}

// MakePrivate converts ABc to aBc.
// Special cases:
//  * ID becomes id
func MakePrivate(str string) string {
	if len(str) == 0 {
		return str
	}

	switch str {
	case "ID":
		return "id"
	default:
		return string(unicode.ToLower(rune(str[0]))) + str[1:]
	}
}
