// Package osc provides a package for sending and receiving OpenSoundControl
// messages. The package is implemented in pure Go.
package osc

import (
	"errors"
	"net"
	"regexp"
	"time"
)

// Dispatcher is an interface for an OSC message dispatcher. A dispatcher is
// responsible for dispatching received OSC messages.
type Dispatcher interface {
	Dispatch(packet Packet, addr net.Addr)
}

// Handler is an interface for message handlers. Every handler implementation
// for an OSC message must implement this interface.
type Handler interface {
	HandleMessage(msg *Message, addr net.Addr)
}

// HandlerFunc implements the Handler interface. Type definition for an OSC
// handler function.
type HandlerFunc func(msg *Message, addr net.Addr)

// HandleMessage calls itself with the given OSC Message. Implements the
// Handler interface.
func (f HandlerFunc) HandleMessage(msg *Message, addr net.Addr) {
	f(msg, addr)
}

////
// StandardDispatcher
////

// StandardDispatcher is a dispatcher for OSC packets. It handles the dispatching of
// received OSC packets to Handlers for their given address.
type StandardDispatcher struct {
	handlers       map[string]Handler
	defaultHandler Handler
}

// NewStandardDispatcher returns an StandardDispatcher.
func NewStandardDispatcher() *StandardDispatcher {
	return &StandardDispatcher{handlers: make(map[string]Handler)}
}

// AddMsgHandler adds a new message handler for the given OSC address.
func (s *StandardDispatcher) AddMsgHandler(oscAddr string, handler HandlerFunc) error {
	// "*" is special
	if oscAddr == "*" {
		s.defaultHandler = handler
		return nil
	}
	// addr is a regex, used to match the input string
	_, err := regexp.Compile(oscAddr)
	if err != nil {
		return err
	}

	if addressExists(oscAddr, s.handlers) {
		return errors.New("OSC address exists already")
	}

	s.handlers[oscAddr] = handler
	return nil
}

// Dispatch dispatches OSC packets. Implements the Dispatcher interface.
func (s *StandardDispatcher) Dispatch(packet Packet, addr net.Addr) {
	switch p := packet.(type) {
	default:
		return

	case *Message:
		for oscAddr, handler := range s.handlers {
			if p.Match(oscAddr) {
				handler.HandleMessage(p, addr)
			}
		}
		if s.defaultHandler != nil {
			s.defaultHandler.HandleMessage(p, addr)
		}

	case *Bundle:
		timer := time.NewTimer(p.Timetag.ExpiresIn())

		go func() {
			<-timer.C
			for _, message := range p.Messages {
				for address, handler := range s.handlers {
					if message.Match(address) {
						handler.HandleMessage(message, addr)
					}
				}
				if s.defaultHandler != nil {
					s.defaultHandler.HandleMessage(message, addr)
				}
			}

			// Process all bundles
			for _, b := range p.Bundles {
				s.Dispatch(b, addr)
			}
		}()
	}
}

// addressExists returns true if the OSC address `oscAddr` is found in `handlers`.
func addressExists(oscAddr string, handlers map[string]Handler) bool {
	for h := range handlers {
		if h == oscAddr {
			return true
		}
	}
	return false
}
