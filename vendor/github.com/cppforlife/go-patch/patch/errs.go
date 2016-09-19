package patch

import (
	"fmt"
)

type opMismatchTypeErr struct {
	type_ string
	path  Pointer
	obj   interface{}
}

func newOpArrayMismatchTypeErr(tokens []Token, obj interface{}) opMismatchTypeErr {
	return opMismatchTypeErr{"an array", NewPointer(tokens), obj}
}

func newOpMapMismatchTypeErr(tokens []Token, obj interface{}) opMismatchTypeErr {
	return opMismatchTypeErr{"a map", NewPointer(tokens), obj}
}

func (e opMismatchTypeErr) Error() string {
	errMsg := "Expected to find %s at path '%s' but found '%T'"
	return fmt.Sprintf(errMsg, e.type_, e.path, e.obj)
}
