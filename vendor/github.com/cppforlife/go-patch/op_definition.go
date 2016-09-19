package patch

import (
	"fmt"
)

// OpDefinition struct is useful for JSON and YAML unmarshaling
type OpDefinition struct {
	Type  string
	Path  *string
	Value *interface{}
}

func NewOpsFromDefinitions(opDefs []OpDefinition) (Ops, error) {
	var ops []Op
	var op Op
	var err error

	for i, opDef := range opDefs {
		switch opDef.Type {
		case "replace":
			op, err = newReplaceOp(opDef)
			if err != nil {
				return nil, fmt.Errorf("Replace operation [%d]: %s", i, err)
			}

		case "remove":
			op, err = newRemoveOp(opDef)
			if err != nil {
				return nil, fmt.Errorf("Remove operation [%d]: %s", i, err)
			}

		default:
			return nil, fmt.Errorf("Unknown operation [%d] with type '%s'", i, opDef.Type)
		}

		ops = append(ops, op)
	}

	return Ops(ops), nil
}

func newReplaceOp(opDef OpDefinition) (ReplaceOp, error) {
	if opDef.Path == nil {
		return ReplaceOp{}, fmt.Errorf("Missing path")
	}

	if opDef.Value == nil {
		return ReplaceOp{}, fmt.Errorf("Missing value")
	}

	ptr, err := NewPointerFromString(*opDef.Path)
	if err != nil {
		return ReplaceOp{}, fmt.Errorf("Invalid path: %s", err)
	}

	return ReplaceOp{Path: ptr, Value: *opDef.Value}, nil
}

func newRemoveOp(opDef OpDefinition) (RemoveOp, error) {
	if opDef.Path == nil {
		return RemoveOp{}, fmt.Errorf("Missing path")
	}

	if opDef.Value != nil {
		return RemoveOp{}, fmt.Errorf("Cannot specify value")
	}

	ptr, err := NewPointerFromString(*opDef.Path)
	if err != nil {
		return RemoveOp{}, fmt.Errorf("Invalid path: %s", err)
	}

	return RemoveOp{Path: ptr}, nil
}
