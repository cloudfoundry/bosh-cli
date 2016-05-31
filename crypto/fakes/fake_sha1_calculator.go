package fakes

import (
	"fmt"
)

type FakeSha1Calculator struct {
	calculateInputs       map[string]CalculateInput
	CalculateStringInputs map[string]string
}

func NewFakeSha1Calculator() *FakeSha1Calculator {
	return &FakeSha1Calculator{}
}

type CalculateInput struct {
	Sha1 string
	Err  error
}

func (c *FakeSha1Calculator) Calculate(path string) (string, error) {
	calculateInput := c.calculateInputs[path]
	return calculateInput.Sha1, calculateInput.Err
}

func (c *FakeSha1Calculator) CalculateString(data string) string {
	if sha1, found := c.CalculateStringInputs[data]; found {
		return sha1
	}
	panic(fmt.Sprintf("Did not find SHA1 result for '%s'", data))
}

func (c *FakeSha1Calculator) SetCalculateBehavior(calculateInputs map[string]CalculateInput) {
	c.calculateInputs = calculateInputs
}
