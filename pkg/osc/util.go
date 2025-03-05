// Package osc provides a package for sending and receiving OpenSoundControl
// messages. The package is implemented in pure Go.
package osc

import (
	"fmt"
)

////
// Utility and helper functions
////

// getTypeTag returns the OSC type tag for the given argument.
func getTypeTag(arg interface{}) (string, error) {
	switch t := arg.(type) {
	case bool:
		if arg.(bool) {
			return "T", nil
		}
		return "F", nil
	case nil:
		return "N", nil
	case int32:
		return "i", nil
	case float32:
		return "f", nil
	case string:
		return "s", nil
	case []byte:
		return "b", nil
	case int64:
		return "h", nil
	case float64:
		return "d", nil
	case Timetag:
		return "t", nil
	default:
		return "", fmt.Errorf("unsupported type: %T", t)
	}
}
