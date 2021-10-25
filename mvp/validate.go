package mvp

import (
	"errors"
	"fmt"
	"github.com/golangee/dyml/token"
	"strings"
)

type Validator interface {
	Validate(domain *Domain) error
}

func (d *Domain) Validate(domain *Domain) error {
	if err := d.ArcVersion.Validate(domain); err != nil {
		return err
	}

	if len(d.Executables) != 1 {
		return token.NewPosError(d.Name.Range, fmt.Sprintf("there must be exactly one executable defined for this domain, but found %d", len(d.Executables)))
	}

	for _, exe := range d.Executables {
		if err := exe.Validate(domain); err != nil {
			return err
		}
	}

	for _, bc := range d.BoundedContexts {
		if err := bc.Validate(domain); err != nil {
			return err
		}
	}

	return nil
}

func (b *BoundedContext) Validate(domain *Domain) error {
	for _, typeImport := range b.Imports {
		if typeImport.Go == nil {
			return fmt.Errorf("type import '%s' needs a go block", typeImport.Name)
		}
	}

	return nil
}

func (e *Executable) Validate(domain *Domain) error {
	if e.Architecture.Type.Value != "4layer" {
		return token.NewPosError(e.Architecture.Type.Range, "architecture type must be '4layer'")
	}

	if err := e.License.Validate(domain); err != nil {
		return err
	}

	return e.Generators.Validate(domain)
}

func (g *GeneratorSelection) Validate(domain *Domain) error {
	if g.Go == nil {
		return errors.New("executable must have a Go generator")
	}

	return nil
}

func (a *ArcVersion) Validate(domain *Domain) error {
	if strings.TrimSpace(a.V.Value) != "0.0.1" {
		return token.NewPosError(a.V.Range, "version must be 0.0.1 for this prototype")
	}

	return nil
}

func (l *License) Validate(domain *Domain) error {
	_, err := l.GetText()
	return err
}
