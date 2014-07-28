package main

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
)

func main() {
	fmt.Println("BOSH Micro CLI")
	bosherr.New("An error")
}
