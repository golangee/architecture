package mvp

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"os"
)

const DefaultFileMode = 0755

// Render creates a project that's ready to run for this domain into the given folder.
func (e *Executable) Render(domain *Domain, folder string) error {
	err := os.MkdirAll(folder, DefaultFileMode)
	if err != nil {
		return err
	}

	f := jen.NewFile("main")

	f.Func().Id("main").Params().Block(
		jen.Id("a").Op(":=").Lit(5),
		jen.Qual("fmt", "Println").Call(jen.Id("a")),
	)
	fmt.Printf("\n\n%#v\n\n", f)

	return nil
}
