package fakes

import "fmt"

type FakeScriptRunner struct {
	RunScripts      []string
	RunScriptErr    error
	runScriptErrors map[string]error
}

func NewFakeScriptRunner() *FakeScriptRunner {
	return &FakeScriptRunner{
		RunScripts:      []string{},
		runScriptErrors: map[string]error{},
	}
}

func (r *FakeScriptRunner) Run(script string) (string, string, error) {
	r.RunScripts = append(r.RunScripts, script)
	if err := r.runScriptErrors[script]; err != nil {
		return "", "", err
	}
	return "", "", r.RunScriptErr
}

func (r *FakeScriptRunner) RegisterRunScriptError(script string, err error) {
	if _, ok := r.runScriptErrors[script]; ok {
		panic(fmt.Sprintf("RunScript error is already set for command: %s", script))
	}
	r.runScriptErrors[script] = err
}
