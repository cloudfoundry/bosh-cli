package completion

import (
	"bytes"
	"reflect"
	"strings"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/spf13/cobra"

	boshcmd "github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
)

const initCmdName = "help"
const cobraCompletionCmdName = "completion"
const cobraCompleteCmdName = "__complete"

const logTag = "completion"

func IsItCompletionCommand(args []string) bool {
	return len(args) > 0 && (args[0] == cobraCompletionCmdName || args[0] == cobraCompleteCmdName)
}

type CapturedResult struct {
	Lines   []string
	Command *cobra.Command
}

type CmdContext struct {
	ConfigPath      string
	EnvironmentName string
	DeploymentName  string
}

func (c *CmdContext) findStoreForFlagValue(flagName string) *string {
	switch flagName {
	case "environment":
		return &c.EnvironmentName
	case "config":
		return &c.ConfigPath
	case "deployment":
		return &c.DeploymentName
	}
	return nil
}

type BoshComplete struct {
	completionFunctionsMap *CompleteFunctionsMap
	rootCmd                *cobra.Command
	cmdContext             *CmdContext
	logger                 boshlog.Logger
}

func NewBoshComplete(blindUi *boshui.ConfUI, logger boshlog.Logger) *BoshComplete {
	deps := boshcmd.NewBasicDeps(blindUi, logger)
	cmdFactory := boshcmd.NewFactory(deps)
	var session boshcmd.Session
	cmd, err := cmdFactory.New([]string{initCmdName}) // just to init session
	if err != nil {
		logger.Debug(logTag, "session initialization (command '%s') error: %v", initCmdName, err)
	} else {
		session = NewCmdBridge(cmd, deps).Session()
	}
	cmdContext := &CmdContext{}
	dq := NewDirectorQuery(logger, cmdContext, session)
	cfMap := NewCompleteFunctionsMap(logger, dq)
	return NewBoshCompleteWithFunctions(logger, cmdContext, cfMap)
}

func NewBoshCompleteWithFunctions(logger boshlog.Logger, cmdContext *CmdContext, completionFunctionsMap *CompleteFunctionsMap) *BoshComplete {
	// https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md

	c := &BoshComplete{
		completionFunctionsMap: completionFunctionsMap,
		logger:                 logger,
		rootCmd:                &cobra.Command{Use: "bosh"},
		cmdContext:             cmdContext,
	}
	c.discoverBoshCommands(c.rootCmd, reflect.TypeOf(opts.BoshOpts{}), 0)
	return c
}

func (c *BoshComplete) discoverBoshCommands(parentCommand *cobra.Command, fieldType reflect.Type, level int) {
	for i := 0; i < fieldType.NumField(); i++ {
		field := fieldType.Field(i)
		if field.Name == "Args" {
			c.tryToBindValidArgsFunction(parentCommand, field.Type.Name())
		} else if field.Tag.Get("long") != "" {
			c.addFlag(parentCommand, field, level == 0)
		} else if fc := field.Tag.Get("command"); fc != "" && fc != cobraCompletionCmdName && fc != cobraCompleteCmdName {
			newCmd := c.addCommand(parentCommand, field)
			if field.Type.Kind() == reflect.Struct {
				c.discoverBoshCommands(newCmd, field.Type, level+1)
			}
		}
	}
}

func (c *BoshComplete) addCommand(parentCommand *cobra.Command, field reflect.StructField) *cobra.Command {
	cmdName := field.Tag.Get("command")
	newCmd := &cobra.Command{
		Use:     cmdName,
		Short:   field.Tag.Get("description"),
		Aliases: c.getTagValues(field.Tag, "alias"),
		Run:     func(_ *cobra.Command, _ []string) {},
	}
	parentCommand.AddCommand(newCmd)
	return newCmd
}

func (c *BoshComplete) getTagValues(fieldTag reflect.StructTag, tagName string) []string {
	rawTag := string(fieldTag)
	parts := strings.Split(rawTag, " ")
	prefix := tagName + ":"
	var values []string
	for _, part := range parts {
		if strings.HasPrefix(part, prefix) {
			value := strings.TrimPrefix(part, prefix)
			values = append(values, value)
		}
	}
	return values
}

func (c *BoshComplete) addFlag(cmd *cobra.Command, field reflect.StructField, rootLevel bool) {
	name := field.Tag.Get("long")
	short := field.Tag.Get("short")
	value := field.Tag.Get("default")
	usage := field.Tag.Get("description")
	env := field.Tag.Get("env")
	if env != "" {
		usage = usage + ", env: " + env
	}
	flagSet := cmd.Flags()
	if rootLevel {
		flagSet = cmd.PersistentFlags()
	}
	p := c.cmdContext.findStoreForFlagValue(name)
	if p == nil {
		flagSet.StringP(name, short, value, usage)
	} else {
		flagSet.StringVarP(p, name, short, value, usage)
	}
	if fun, ok := (*c.completionFunctionsMap)["--"+name]; ok {
		err := cmd.RegisterFlagCompletionFunc(name, fun)
		if err != nil {
			c.logger.Warn(logTag, "register flag %s completion function error: %v", name, err)
		}
	}
}

func (c *BoshComplete) Execute(args []string) error {
	c.rootCmd.SetArgs(args)
	_, err := c.rootCmd.ExecuteC()
	return err
}

func (c *BoshComplete) ExecuteCaptured(args []string) (*CapturedResult, error) {
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	c.rootCmd.SetOut(outBuf)
	c.rootCmd.SetErr(errBuf)
	c.rootCmd.SetArgs(args)
	retCmd, err := c.rootCmd.ExecuteC()
	if err != nil {
		return nil, err
	}
	retLines := strings.Split(outBuf.String(), "\n")
	c.logger.Debug("BoshComplete.ExecuteCapture() STDERR", errBuf.String())
	return &CapturedResult{Lines: retLines, Command: retCmd}, nil
}

func (c *BoshComplete) tryToBindValidArgsFunction(cmd *cobra.Command, argsTypeName string) {
	if fun, ok := (*c.completionFunctionsMap)[argsTypeName]; ok {
		//c.logger.Debug(c.logTag, "Command ValidArgsFunction '%s': `%v`", cmd.Name(), GetShortFunName(fun))
		cmd.ValidArgsFunction = fun
	} else {
		c.logger.Warn(logTag, "Unknown Args Type %s, command %s", argsTypeName, cmd.Name())
	}
}
