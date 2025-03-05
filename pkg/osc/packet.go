// Package osc provides a package for sending and receiving OpenSoundControl
// messages. The package is implemented in pure Go.
package osc

import (
	"bufio"
	"bytes"
	"encoding"
)

// Packet is the interface for Message and Bundle.
type Packet interface {
	encoding.BinaryMarshaler
}

// ParsePacket parses the given msg string and returns a Packet
func ParsePacket(msg string) (Packet, error) {
	p, err := readPacket(bufio.NewReader(bytes.NewBufferString(msg)))
	if err != nil {
		return nil, err
	}
	return p, nil
}

// receivePacket receives an OSC packet from the given reader.
func readPacket(reader *bufio.Reader) (Packet, error) {
	//var buf []byte
	buf, err := reader.Peek(1)
	if err != nil {
		return nil, err
	}

	// An OSC Message starts with a '/'
	if buf[0] == '/' {
		packet, err := readMessage(reader)
		if err != nil {
			return nil, err
		}
		return packet, nil
	}
	if buf[0] == '#' { // An OSC bundle starts with a '#'
		packet, err := readBundle(reader)
		if err != nil {
			return nil, err
		}
		return packet, nil
	}

	var p Packet
	return p, nil
}
