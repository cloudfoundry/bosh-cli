package cmd

type Cmd interface {
	Run([]string) error
}
