// Package osc provides a package for sending and receiving OpenSoundControl
// messages. The package is implemented in pure Go.
package osc

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

const (
	bundleTagString = "#bundle"
)

// Bundle represents an OSC bundle. It consists of the OSC-string "#bundle"
// followed by an OSC Time Tag, followed by zero or more OSC bundle/message
// elements. The OSC-timetag is a 64-bit fixed point time tag. See
// http://opensoundcontrol.org/spec-1_0 for more information.
type Bundle struct {
	Timetag  Timetag
	Messages []*Message
	Bundles  []*Bundle
}

// Verify that Bundle implements the Packet interface.
var _ Packet = (*Bundle)(nil)

////
// Bundle
////

// NewBundle returns an OSC Bundle. Use this function to create a new OSC
// Bundle.
func NewBundle(time time.Time) *Bundle {
	return &Bundle{Timetag: *NewTimetag(time)}
}

// Append appends an OSC bundle or OSC message to the bundle.
func (b *Bundle) Append(pck Packet) error {
	switch t := pck.(type) {
	default:
		return fmt.Errorf("unsupported OSC packet type: only Bundle and Message are supported")

	case *Bundle:
		b.Bundles = append(b.Bundles, t)

	case *Message:
		b.Messages = append(b.Messages, t)
	}

	return nil
}

// MarshalBinary serializes the OSC bundle to a byte array with the following
// format:
// 1. Bundle string: '#bundle'
// 2. OSC timetag
// 3. Length of first OSC bundle element
// 4. First bundle element
// 5. Length of n OSC bundle element
// 6. n bundle element
func (b *Bundle) MarshalBinary() ([]byte, error) {
	// Add the '#bundle' string
	data := new(bytes.Buffer)
	if _, err := writePaddedString("#bundle", data); err != nil {
		return nil, err
	}

	// Add the time tag
	bd, err := b.Timetag.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if _, err = data.Write(bd); err != nil {
		return nil, err
	}

	// Process all OSC Messages
	for _, m := range b.Messages {
		buf, err := m.MarshalBinary()
		if err != nil {
			return nil, err
		}

		// Append the length of the OSC message
		if err = binary.Write(data, binary.BigEndian, int32(len(buf))); err != nil {
			return nil, err
		}

		// Append the OSC message
		if _, err = data.Write(buf); err != nil {
			return nil, err
		}
	}

	// Process all OSC Bundles
	for _, b := range b.Bundles {
		buf, err := b.MarshalBinary()
		if err != nil {
			return nil, err
		}

		// Write the size of the bundle
		if err = binary.Write(data, binary.BigEndian, int32(len(buf))); err != nil {
			return nil, err
		}

		// Append the bundle
		_, err = data.Write(buf)
		if err != nil {
			return nil, err
		}
	}

	return data.Bytes(), nil
}

// readBundle reads an Bundle from reader.
func readBundle(reader *bufio.Reader) (*Bundle, error) {
	// Read the '#bundle' OSC string
	startTag, _, err := readPaddedString(reader)
	if err != nil {
		return nil, err
	}

	if startTag != bundleTagString {
		return nil, fmt.Errorf("invalid bundle start tag: %s", startTag)
	}

	// Read the timetag
	var timeTag uint64
	if err := binary.Read(reader, binary.BigEndian, &timeTag); err != nil {
		return nil, err
	}

	// Create a new bundle
	bundle := NewBundle(timetagToTime(timeTag))

	// Read until the end of the buffer
	for reader.Buffered() > 0 {
		// Read the size of the bundle element
		var length int32
		if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
			return nil, err
		}

		p, err := readPacket(reader)
		if err != nil {
			return nil, err
		}
		if err = bundle.Append(p); err != nil {
			return nil, err
		}
	}

	return bundle, nil
}
