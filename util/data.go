package util

import (
	"errors"
	"fmt"
)

// TraversePath takes a JSON-like value and recursively traverses it with a
// given list of keys. The result is a subset of that input value or an error
func TraversePath(data interface{}, path *[]interface{}) (interface{}, error) {
	var child interface{} = data
	var ok bool

	for _, segment := range *path {
		switch v := child.(type) {
		case map[string]interface{}:
			switch i := segment.(type) {
			case string:
				child, ok = v[i]
				if !ok {
					return nil, makeMissingDataError("Could not traverse data path, object has no property at given name")
				}

			case int:
				child, ok = v[fmt.Sprint(i)]
				if !ok {
					return nil, makeMissingDataError("Could not traverse data path, object has no property at given name")
				}

			default:
				return nil, errors.New("Could not traverse data path, tried to index map with non-int, non-string value")
			}

		case []interface{}:
			switch i := segment.(type) {
			case int:
				if i < len(v) {
					child = v[i]
				} else {
					return nil, makeMissingDataError("Could not traverse data path, array has no property at index")
				}

			default:
				return nil, errors.New("Could not traverse data path, tried to index array with non-int type")
			}

		default:
			return nil, errors.New("Could not traverse data path, tried to index into a non-array, non-object value")
		}
	}

	return child, nil
}

// SetAtPath is the write equivalent of TraversePath. It will write a subset
// of the given value with the given new data at the provided path. It will also
// create maps and arrays to reach the provided path if they don't exist, as
// well as filling out arrays with null values to make them long enough to fit
// a given path segment.
func SetAtPath(data *interface{}, path *[]interface{}, newData interface{}) error {
	if path == nil || len(*path) == 0 {
		*data = newData
	} else {
		var child interface{} = *data
		var lastParent interface{}
		var ok bool

		lastIdx := len(*path) - 1

		for segmentIdx, segment := range *path {
			switch v := child.(type) {
			case map[string]interface{}:
				switch i := segment.(type) {
				case string:
					if segmentIdx == lastIdx {
						v[i] = newData
					} else {
						child, ok = v[i]
						if !ok {
							child = createChild((*path)[segmentIdx+1])
							v[i] = child
						}
					}

				case int:
					if segmentIdx == lastIdx {
						v[fmt.Sprint(i)] = newData
					} else {
						child, ok = v[fmt.Sprint(i)]
						if !ok {
							child = createChild((*path)[segmentIdx+1])
							v[fmt.Sprint(i)] = child
						}
					}

				default:
					return errors.New("Could not set data at path, tried to index map with non-int, non-string value")
				}

				lastParent = v

			case []interface{}:
				switch i := segment.(type) {
				case int:
					if i < len(v) {
						if segmentIdx == lastIdx {
							v[i] = newData
						} else {
							child = v[i]
						}
					} else {
						// We need to insert spaces in the slice, and then
						// overwrite the previous slice wherever it was stored
						gap := make([]interface{}, i-len(v))
						v = append(v, gap...)

						if segmentIdx == lastIdx {
							v = append(v, newData)
						} else {
							child = createChild((*path)[segmentIdx+1])
							v = append(v, child)
						}

						switch l := lastParent.(type) {
						case nil:
							*data = v

						case map[string]interface{}:
							switch li := (*path)[segmentIdx-1].(type) {
							case string:
								l[li] = v

							case int:
								l[fmt.Sprint(li)] = v

							default:
								panic("unreachable")
							}

						case []interface{}:
							switch li := (*path)[segmentIdx-1].(type) {
							case int:
								l[li] = v

							default:
								panic("unreachable")
							}

						default:
							panic("unreachable")
						}
					}
				}

				lastParent = v

			default:
				return errors.New("Could not set data at path, tried to index into a non-array, non-object value")
			}
		}
	}

	return nil
}

func createChild(nextSegment interface{}) interface{} {
	switch nextSegment.(type) {
	case int:
		return []interface{}{}

	case string:
		return make(map[string]interface{})

	default:
		panic("Unexpected path segment, not an int or string")
	}
}

type missingDataError struct {
	message string
}

func (err *missingDataError) Error() string {
	return err.message
}

func makeMissingDataError(message string) error {
	return &missingDataError{
		message: message,
	}
}

// IsMissingData returns true if the error is a missingDataError
func IsMissingData(err error) bool {
	switch err.(type) {
	case *missingDataError:
		return true

	default:
		return false
	}
}
