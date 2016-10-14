package patch

import (
	"fmt"
)

type ReplaceOp struct {
	Path  Pointer
	Value interface{}
}

func (op ReplaceOp) Apply(doc interface{}) (interface{}, error) {
	tokens := op.Path.Tokens()

	if len(tokens) == 1 {
		return op.Value, nil
	}

	obj := doc
	prevUpdate := func(newObj interface{}) { doc = newObj }

	for i, token := range tokens[1:] {
		isLast := i == len(tokens)-2

		switch typedToken := token.(type) {
		case IndexToken:
			idx := typedToken.Index

			typedObj, ok := obj.([]interface{})
			if !ok {
				return nil, newOpArrayMismatchTypeErr(tokens[:i+2], obj)
			}

			if idx >= len(typedObj) {
				errMsg := "Expected to find array index '%d' but found array of length '%d'"
				return nil, fmt.Errorf(errMsg, idx, len(typedObj))
			}

			if isLast {
				typedObj[idx] = op.Value
			} else {
				obj = typedObj[idx]
				prevUpdate = func(newObj interface{}) { typedObj[idx] = newObj }
			}

		case AfterLastIndexToken:
			typedObj, ok := obj.([]interface{})
			if !ok {
				return nil, newOpArrayMismatchTypeErr(tokens[:i+2], obj)
			}

			if isLast {
				prevUpdate(append(typedObj, op.Value))
			} else {
				return nil, fmt.Errorf("Expected after last index token to be last in path '%s'", op.Path)
			}

		case MatchingIndexToken:
			typedObj, ok := obj.([]interface{})
			if !ok {
				return nil, newOpArrayMismatchTypeErr(tokens[:i+2], obj)
			}

			var idxs []int

			for itemIdx, item := range typedObj {
				typedItem, ok := item.(map[interface{}]interface{})
				if ok {
					if typedItem[typedToken.Key] == typedToken.Value {
						idxs = append(idxs, itemIdx)
					}
				}
			}

			if typedToken.Optional && len(idxs) == 0 {
				obj = map[interface{}]interface{}{typedToken.Key: typedToken.Value}
				prevUpdate(append(typedObj, obj))
				// no need to change prevUpdate since matching item can only be a map
			} else {
				if len(idxs) != 1 {
					errMsg := "Expected to find exactly one matching array item for path '%s' but found %d"
					return nil, fmt.Errorf(errMsg, NewPointer(tokens[:i+2]), len(idxs))
				}

				idx := idxs[0]

				if isLast {
					typedObj[idx] = op.Value
				} else {
					obj = typedObj[idx]
					// no need to change prevUpdate since matching item can only be a map
				}
			}

		case KeyToken:
			typedObj, ok := obj.(map[interface{}]interface{})
			if !ok {
				return nil, newOpMapMismatchTypeErr(tokens[:i+2], obj)
			}

			var found bool

			obj, found = typedObj[typedToken.Key]
			if !found && !typedToken.Optional {
				errMsg := "Expected to find a map key '%s' for path '%s'"
				return nil, fmt.Errorf(errMsg, typedToken.Key, NewPointer(tokens[:i+2]))
			}

			if isLast {
				typedObj[typedToken.Key] = op.Value
			} else {
				prevUpdate = func(newObj interface{}) { typedObj[typedToken.Key] = newObj }

				if !found {
					// Determine what type of value to create based on next token
					switch tokens[i+2].(type) {
					case AfterLastIndexToken:
						obj = []interface{}{}
					case MatchingIndexToken:
						obj = []interface{}{}
					case KeyToken:
						obj = map[interface{}]interface{}{}
					default:
						errMsg := "Expected to find key, matching index or after last index token at path '%s'"
						return nil, fmt.Errorf(errMsg, NewPointer(tokens[:i+3]))
					}

					typedObj[typedToken.Key] = obj
				}
			}

		default:
			return nil, fmt.Errorf("Expected to not find token '%T' at '%s'", token, NewPointer(tokens[:i+2]))
		}
	}

	return doc, nil
}
