package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"

	// Should only be imported here to avoid leaking use of goflags through project
	goflags "github.com/jessevdk/go-flags"

	bconfigserver "github.com/cloudfoundry/bosh-cli/configserver"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	fsext "github.com/cloudfoundry/bosh-cli/fsext"
)

type Factory struct {
	deps BasicDeps
}

func NewFactory(deps BasicDeps) Factory {
	return Factory{deps: deps}
}

func (f Factory) New(args []string) (Cmd, error) {
	var cmdOpts interface{}

	boshOpts := &BoshOpts{}

	boshOpts.VersionOpt = func() error {
		return &goflags.Error{
			Type:    goflags.ErrHelp,
			Message: fmt.Sprintf("version %s\n", VersionLabel),
		}
	}

	parser := goflags.NewParser(boshOpts, goflags.HelpFlag|goflags.PassDoubleDash)

	for _, c := range parser.Commands() {
		docsURL := "https://bosh.io/docs/cli-v2#" + c.Name

		c.LongDescription = c.ShortDescription + "\n\n" + docsURL

		fillerLen := 50 - len(c.ShortDescription)
		if fillerLen < 0 {
			fillerLen = 0
		}

		c.ShortDescription += strings.Repeat(" ", fillerLen+1) + docsURL
	}

	schemaFS := fsext.NewSchemaDelegatingFS(f.deps.FS)
	schemaFS.RegisterSchema("config-server", bconfigserver.NewFS(bconfigserver.NewErrClient()))
	schemaFS.RegisterSchema("memory", fsext.NewInMemoryFS())

	f.deps.FS = schemaFS

	parser.CommandHandler = func(command goflags.Commander, extraArgs []string) error {
		if opts, ok := command.(*SSHOpts); ok {
			if len(opts.Command) == 0 {
				opts.Command = extraArgs
				extraArgs = []string{}
			}
		}

		if opts, ok := command.(*AliasEnvOpts); ok {
			opts.URL = boshOpts.EnvironmentOpt
			opts.CACert = boshOpts.CACertOpt
		}

		if opts, ok := command.(*EventsOpts); ok {
			opts.Deployment = boshOpts.DeploymentOpt
		}

		if opts, ok := command.(*VMsOpts); ok {
			opts.Deployment = boshOpts.DeploymentOpt
		}

		if opts, ok := command.(*InstancesOpts); ok {
			opts.Deployment = boshOpts.DeploymentOpt
		}

		if opts, ok := command.(*TasksOpts); ok {
			opts.Deployment = boshOpts.DeploymentOpt
		}

		if opts, ok := command.(*TaskOpts); ok {
			opts.Deployment = boshOpts.DeploymentOpt
		}

		var client bconfigserver.Client

		if len(boshOpts.ConfigServerFlags.URL) > 0 {
			clientOpts := bconfigserver.HTTPClientOpts{
				URL:            boshOpts.ConfigServerFlags.URL,
				TLSCA:          []byte(boshOpts.ConfigServerFlags.TLSCA.Content),
				TLSCertificate: []byte(boshOpts.ConfigServerFlags.TLSCertificate.Content),
				TLSPrivateKey:  []byte(boshOpts.ConfigServerFlags.TLSPrivateKey.Content),
				Namespace:      boshOpts.ConfigServerFlags.Namespace,
			}

			var err error

			client, err = bconfigserver.NewHTTPClient(clientOpts, f.deps.Logger)
			if err != nil {
				return fmt.Errorf("Failed to build config server HTTP client: %s", err)
			}
		} else {
			client = bconfigserver.NewErrClient()
		}

		schemaFS.RegisterSchema("config-server", bconfigserver.NewFS(client))

		varsStores := map[string]boshtpl.Variables{
			"config-server": NewConfigServerVarsStore(client),
		}

		switch opts := command.(type) {
		case *CreateEnvOpts:
			opts.VarFlags.VarsStore.RegisterSchemas(varsStores)
		case *DeleteEnvOpts:
			opts.VarFlags.VarsStore.RegisterSchemas(varsStores)
		case *InterpolateOpts:
			opts.VarFlags.VarsStore.RegisterSchemas(varsStores)
		}

		if len(extraArgs) > 0 {
			errMsg := "Command '%T' does not support extra arguments: %s"
			return fmt.Errorf(errMsg, command, strings.Join(extraArgs, ", "))
		}

		cmdOpts = command

		return nil
	}

	boshOpts.SSH.GatewayFlags.UUIDGen = f.deps.UUIDGen
	boshOpts.SCP.GatewayFlags.UUIDGen = f.deps.UUIDGen
	boshOpts.Logs.GatewayFlags.UUIDGen = f.deps.UUIDGen

	goflags.FactoryFunc = func(val interface{}) {
		stype := reflect.Indirect(reflect.ValueOf(val))
		if stype.Kind() == reflect.Struct {
			field := stype.FieldByName("FS")
			if field.IsValid() {
				field.Set(reflect.ValueOf(f.deps.FS))
			}
		}
	}

	helpText := bytes.NewBufferString("")
	parser.WriteHelp(helpText)

	_, err := parser.ParseArgs(args)

	if boshOpts.UsernameOpt != "" {
		return Cmd{}, errors.New("BOSH_USER is deprecated use BOSH_CLIENT instead")
	}

	// --help and --version result in errors; turn them into successful output cmds
	if typedErr, ok := err.(*goflags.Error); ok {
		if typedErr.Type == goflags.ErrHelp {
			cmdOpts = &MessageOpts{Message: typedErr.Message}
			err = nil
		}
	}

	if _, ok := cmdOpts.(*HelpOpts); ok {
		cmdOpts = &MessageOpts{Message: helpText.String()}
	}

	return NewCmd(*boshOpts, cmdOpts, f.deps), err
}
