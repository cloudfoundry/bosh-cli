package cmd

import (
	"fmt"
	"reflect"
	"strings"

	// Should only be imported here to avoid leaking use of goflags through project
	goflags "github.com/jessevdk/go-flags"
)

type Factory struct {
	deps BasicDeps
}

func NewFactory(deps BasicDeps) Factory {
	return Factory{deps: deps}
}

func (f Factory) New(args []string) (Cmd, error) {
	cmd := NewCmd(&BoshOpts{}, nil, f.deps)

	cmd.BoshOpts.VersionOpt = func() error {
		return &goflags.Error{
			Type:    goflags.ErrHelp,
			Message: fmt.Sprintf("version %s", VersionLabel),
		}
	}

	parser := goflags.NewParser(cmd.BoshOpts, goflags.HelpFlag|goflags.PassDoubleDash)

	parser.CommandHandler = func(command goflags.Commander, extraArgs []string) error {
		if opts, ok := command.(*SSHOpts); ok {
			if len(opts.Command) == 0 {
				opts.Command = extraArgs
				extraArgs = []string{}
			}
		}

		if opts, ok := command.(*EnvironmentOpts); ok {
			opts.CACert = cmd.BoshOpts.CACertOpt
		}

		if opts, ok := command.(*EventsOpts); ok {
			opts.Deployment = cmd.BoshOpts.DeploymentOpt
		}

		if len(extraArgs) > 0 {
			errMsg := "Command '%T' does not support extra arguments: %s"
			return fmt.Errorf(errMsg, command, strings.Join(extraArgs, ", "))
		}

		cmd.Opts = command

		return nil
	}

	goflags.FactoryFunc = func(val interface{}) {
		stype := reflect.Indirect(reflect.ValueOf(val))
		if stype.Kind() == reflect.Struct {
			field := stype.FieldByName("FS")
			if field.IsValid() {
				field.Set(reflect.ValueOf(f.deps.FS))
			}
		}
	}

	_, err := parser.ParseArgs(args)

	return cmd, err
}
