// Package osc provides a package for sending and receiving OpenSoundControl
// messages. The package is implemented in pure Go.
package osc

import (
	"bufio"
	"bytes"
	"net"
	"strings"
	"time"
)

// Server represents an OSC server. The server listens on Address and Port for
// incoming OSC packets and bundles.
type Server struct {
	Addr        string
	Dispatcher  Dispatcher
	ReadTimeout time.Duration
	close       func() error
}

// ListenAndServe retrieves incoming OSC packets and dispatches the retrieved
// OSC packets.
func (s *Server) ListenAndServe() error {
	defer s.CloseConnection()

	if s.Dispatcher == nil {
		s.Dispatcher = NewStandardDispatcher()
	}

	ln, err := net.ListenPacket("udp", s.Addr)
	if err != nil {
		return err
	}

	s.close = ln.Close

	return s.Serve(ln)
}

// Serve retrieves incoming OSC packets from the given connection and dispatches
// retrieved OSC packets. If something goes wrong an error is returned.
func (s *Server) Serve(c net.PacketConn) error {
	var tempDelay time.Duration
	for {
		msg, addr, err := s.readFromConnection(c)
		if err != nil {
			// This was looking at ne.Temporary() which is deprecated
			if _, ok := err.(net.Error); ok {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			return err
		}
		tempDelay = 0
		go s.Dispatcher.Dispatch(msg, addr)
	}
}

// CloseConnection forcibly closes a server's connection.
//
// This causes a "use of closed network connection" error the next time the
// server attempts to read from the connection.
func (s *Server) CloseConnection() error {
	if s.close == nil {
		return nil
	}

	err := s.close()
	// If we get "use of closed network connection", it's not a problem because
	// closing the network connection is exactly what we wanted to do!
	if err != nil && !strings.Contains(
		err.Error(), "use of closed network connection",
	) {
		return err
	}

	return nil
}

// ReceivePacket listens for incoming OSC packets and returns the packet if one is received.
func (s *Server) ReceivePacket(c net.PacketConn) (Packet, net.Addr, error) {
	return s.readFromConnection(c)
}

// readFromConnection retrieves OSC packets.
func (s *Server) readFromConnection(c net.PacketConn) (Packet, net.Addr, error) {
	if s.ReadTimeout != 0 {
		if err := c.SetReadDeadline(time.Now().Add(s.ReadTimeout)); err != nil {
			return nil, nil, err
		}
	}

	data := make([]byte, 65535)
	n, addr, err := c.ReadFrom(data)
	if err != nil {
		return nil, nil, err
	}

	p, err := readPacket(bufio.NewReader(bytes.NewBuffer(data[0:n])))
	if err != nil {
		return nil, nil, err
	}
	return p, addr, nil
}
