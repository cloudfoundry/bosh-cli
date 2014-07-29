package cmd

type Runner struct {
	factory Factory
	args    []string
}

func NewRunner(factory Factory) *Runner {
	return &Runner{factory: factory}
}

func (runner *Runner) Run(args []string) error {
	runner.args = args

	return nil
}
