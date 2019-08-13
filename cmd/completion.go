package cmd

import "fmt"

const script = `_bosh ()
{
  args=("${COMP_WORDS[@]:1:$COMP_CWORD}")
  local IFS=$'\n'

  COMPREPLY=($(GO_FLAGS_COMPLETION=verbose ${COMP_WORDS[0]} "${args[@]}"))
  return 0
}

complete -d -F _bosh bosh
`

type CompletionCmd struct{}

func NewCompletionCmd() CompletionCmd {
	return CompletionCmd{}
}

func (c CompletionCmd) Run(opts CompletionOpts) error {
	fmt.Println(script)

	return nil
}
