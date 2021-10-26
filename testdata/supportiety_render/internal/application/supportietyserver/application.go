// Code generated by golangee/eearc; DO NOT EDIT.

package supportietyserver

import (
	context "context"
	errors "errors"
	flag "flag"
	fmt "fmt"
	os "os"
	filepath "path/filepath"
)

// Application embeds the defaultApplication to provide the default application behavior.
// It also provides the inversion of control injection mechanism for all bounded contexts.
type Application struct {
	defaultApplication
}

func NewApplication(ctx context.Context) (*Application, error) {
	a := &Application{}
	a.defaultApplication.self = a
	if err := a.init(ctx); err != nil {
		return nil, fmt.Errorf("cannot init application: %w", err)
	}

	return a, nil
}

// defaultApplication aggregates all contained bounded contexts and starts their driver adapters.
type defaultApplication struct {
	// cfg contains the global read-only configuration for all bounded contexts.
	cfg Configuration

	// self provides a pointer to the actual Application instance to provide
	// one level of a quasi-vtable calling indirection for simple method 'overriding'.
	self *Application
}

// configure resets, prepares and parses the configuration. The priority of evaluation is:
//
//   0. hardcoded defaults
//   1. values from configuration file
//   2. values from environment variables
//   3. values from command flags
func (d *defaultApplication) configure() error {
	const (
		appName      = "supportiety_server"
		fileFlagHelp = "filename to a configuration file in JSON format."
	)

	// prio 0: hardcoded defaults
	d.cfg.Reset()

	// prio 1: values from configuration file
	usrCfgHome, err := os.UserConfigDir()
	if err == nil {
		usrCfgHome = filepath.Join(usrCfgHome, "."+appName, "settings.json")
	}

	fileFlagSet := flag.NewFlagSet(appName, flag.ContinueOnError)
	d.cfg.ConfigureFlags(fileFlagSet)
	filename := fileFlagSet.String("config", usrCfgHome, fileFlagHelp)
	if err := fileFlagSet.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}

	// note: we now loaded already all flags into the configuration, which is not correct.
	// therefore we do it later once more, to maintain correct order.
	if *filename != "" {
		if err := d.cfg.ParseFile(*filename); err != nil {
			if *filename != usrCfgHome || !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("cannot explicitly parse configuration file: %w", err)
			}
		}
	}

	// prio 2: values from environment variables
	if err := d.cfg.ParseEnv(); err != nil {
		return fmt.Errorf("cannot parse environment variables: %w", err)
	}

	// prio 3: finally parse again the values from the actual command line
	cfgFlagSet := flag.NewFlagSet(appName, flag.ContinueOnError)
	_ = cfgFlagSet.String("config", usrCfgHome, fileFlagHelp) // announce also the config file flag for proper parsing and help
	d.cfg.ConfigureFlags(cfgFlagSet)
	if err := cfgFlagSet.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}

	return nil
}

func (d *defaultApplication) init(ctx context.Context) error {
	if err := d.configure(); err != nil {
		return fmt.Errorf("cannot configure: %w", err)
	}

	return nil
}

func (_ defaultApplication) Run(ctx context.Context) error {
	return nil
}
