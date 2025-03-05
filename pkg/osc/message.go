// Package osc provides a package for sending and receiving OpenSoundControl
// messages. The package is implemented in pure Go.
package osc

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"regexp"
)

// Message represents a single OSC message. An OSC message consists of an OSC
// address pattern and zero or more arguments.
type Message struct {
	Address   string
	Arguments []interface{}
}

// Verify that Messages implements the Packet interface.
var _ Packet = (*Message)(nil)

////
// Message
////

// NewMessage returns a new Message. The address parameter is the OSC address.
func NewMessage(addr string, args ...interface{}) *Message {
	return &Message{Address: addr, Arguments: args}
}

// Append appends the given arguments to the arguments list.
func (msg *Message) Append(args ...interface{}) {
	msg.Arguments = append(msg.Arguments, args...)
}

// Equals returns true if the given OSC Message `m` is equal to the current OSC
// Message. It checks if the OSC address and the arguments are equal. Returns
// true if the current object and `m` are equal.
func (msg *Message) Equals(m *Message) bool {
	return reflect.DeepEqual(msg, m)
}

// Clear clears the OSC address and all arguments.
func (msg *Message) Clear() {
	msg.Address = ""
	msg.ClearData()
}

// ClearData removes all arguments from the OSC Message.
func (msg *Message) ClearData() {
	msg.Arguments = msg.Arguments[len(msg.Arguments):]
}

// Match returns true, if the OSC address pattern of the OSC Message matches the given
// address. The match is case sensitive!
func (msg *Message) Match(addr string) bool {
	exp := regexp.MustCompile(addr)
	return exp.MatchString(msg.Address)
}

// TypeTags returns the type tag string.
func (msg *Message) TypeTags() (string, error) {
	if msg == nil {
		return "", fmt.Errorf("message is nil")
	}

	tags := ","
	for _, m := range msg.Arguments {
		s, err := getTypeTag(m)
		if err != nil {
			return "", err
		}
		tags += s
	}

	return tags, nil
}

// String implements the fmt.Stringer interface.
func (msg *Message) String() string {
	if msg == nil {
		return ""
	}

	tags, err := msg.TypeTags()
	if err != nil {
		return ""
	}

	formatString := "%s %s"
	var args []interface{}
	args = append(args, msg.Address)
	args = append(args, tags)

	for _, arg := range msg.Arguments {
		switch arg := arg.(type) {
		case bool, int32, int64, float32, float64, string:
			formatString += " %v"
			args = append(args, arg)

		case nil:
			formatString += " %s"
			args = append(args, "Nil")

		case []byte:
			formatString += " %s"
			args = append(args, "blob")

		case Timetag:
			formatString += " %d"
			args = append(args, arg.TimeTag())
		}
	}

	return fmt.Sprintf(formatString, args...)
}

// CountArguments returns the number of arguments.
func (msg *Message) CountArguments() int {
	return len(msg.Arguments)
}

// MarshalBinary serializes the OSC message to a byte buffer. The byte buffer
// has the following format:
// 1. OSC Address Pattern
// 2. OSC Type Tag String
// 3. OSC Arguments
func (msg *Message) MarshalBinary() ([]byte, error) {
	// We can start with the OSC address and add it to the buffer
	data := new(bytes.Buffer)
	if _, err := writePaddedString(msg.Address, data); err != nil {
		return nil, err
	}

	// Type tag string starts with ","
	typetags := []byte{','}

	// Process the type tags and collect all arguments
	payload := new(bytes.Buffer)
	for _, arg := range msg.Arguments {
		// FIXME: Use t instead of arg
		switch t := arg.(type) {
		default:
			return nil, fmt.Errorf("OSC - unsupported type: %T", t)

		case bool:
			if arg.(bool) {
				typetags = append(typetags, 'T')
			} else {
				typetags = append(typetags, 'F')
			}

		case nil:
			typetags = append(typetags, 'N')

		case int32:
			typetags = append(typetags, 'i')
			if err := binary.Write(payload, binary.BigEndian, int32(t)); err != nil {
				return nil, err
			}

		case float32:
			typetags = append(typetags, 'f')
			if err := binary.Write(payload, binary.BigEndian, float32(t)); err != nil {
				return nil, err
			}

		case string:
			typetags = append(typetags, 's')
			if _, err := writePaddedString(t, payload); err != nil {
				return nil, err
			}

		case []byte:
			typetags = append(typetags, 'b')
			if _, err := writeBlob(t, payload); err != nil {
				return nil, err
			}

		case int64:
			typetags = append(typetags, 'h')
			if err := binary.Write(payload, binary.BigEndian, int64(t)); err != nil {
				return nil, err
			}

		case float64:
			typetags = append(typetags, 'd')
			if err := binary.Write(payload, binary.BigEndian, float64(t)); err != nil {
				return nil, err
			}

		case Timetag:
			typetags = append(typetags, 't')
			timeTag := arg.(Timetag)
			b, err := timeTag.MarshalBinary()
			if err != nil {
				return nil, err
			}
			if _, err = payload.Write(b); err != nil {
				return nil, err
			}
		}
	}

	// Write the type tag string to the data buffer
	if _, err := writePaddedString(string(typetags), data); err != nil {
		return nil, err
	}

	// Write the payload (OSC arguments) to the data buffer
	if _, err := data.Write(payload.Bytes()); err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}

// readMessage from `reader`.
func readMessage(reader *bufio.Reader) (*Message, error) {
	// First, read the OSC address
	addr, _, err := readPaddedString(reader)
	if err != nil {
		return nil, err
	}

	// Read all arguments
	msg := NewMessage(addr)
	if err = readArguments(msg, reader); err != nil {
		return nil, err
	}

	return msg, nil
}

// readArguments from `reader` and add them to the OSC message `msg`.
func readArguments(msg *Message, reader *bufio.Reader) error {
	// Read the type tag string
	typetags, _, err := readPaddedString(reader)
	if err != nil {
		return err
	}

	if len(typetags) == 0 {
		return nil
	}

	// If the typetag doesn't start with ',', it's not valid
	if typetags[0] != ',' {
		return fmt.Errorf("unsupported type tag string %s", typetags)
	}

	// Remove ',' from the type tag
	typetags = typetags[1:]

	for _, c := range typetags {
		switch c {
		default:
			return fmt.Errorf("unsupported type tag: %c", c)

		case 'i': // int32
			var i int32
			if err = binary.Read(reader, binary.BigEndian, &i); err != nil {
				return err
			}
			msg.Append(i)

		case 'h': // int64
			var i int64
			if err = binary.Read(reader, binary.BigEndian, &i); err != nil {
				return err
			}
			msg.Append(i)

		case 'f': // float32
			var f float32
			if err = binary.Read(reader, binary.BigEndian, &f); err != nil {
				return err
			}
			msg.Append(f)

		case 'd': // float64/double
			var d float64
			if err = binary.Read(reader, binary.BigEndian, &d); err != nil {
				return err
			}
			msg.Append(d)

		case 's': // string
			// TODO: fix reading string value
			var s string
			if s, _, err = readPaddedString(reader); err != nil {
				return err
			}
			msg.Append(s)

		case 'b': // blob
			var buf []byte
			if buf, _, err = readBlob(reader); err != nil {
				return err
			}
			msg.Append(buf)

		case 't': // OSC time tag
			var tt uint64
			if err = binary.Read(reader, binary.BigEndian, &tt); err != nil {
				return nil
			}
			msg.Append(*NewTimetagFromTimetag(tt))

		case 'N': // nil
			msg.Append(nil)

		case 'T': // true
			msg.Append(true)

		case 'F': // false
			msg.Append(false)
		}
	}

	return nil
}

// PrintMessage pretty prints an OSC message to the standard output.
func PrintMessage(msg *Message) {
	fmt.Println(msg)
}
