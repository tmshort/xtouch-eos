// Package osc provides a package for sending and receiving OpenSoundControl
// messages. The package is implemented in pure Go.
package osc

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"
)

type Connection struct {
	laddr       *net.UDPAddr
	raddr       *net.UDPAddr
	conn        *net.UDPConn
	Dispatcher  Dispatcher
	ReadTimeout time.Duration
	errorCh     chan error
	readCh      chan readData
	doneCh      chan bool
}

// NewConnection creates a new OSC client/server. The Connection is used to send and receive OSC
// messages and OSC bundles over an UDP network connection. The `lport` argument specifies the
// local port to list to and send on. The `raddr` argument specifies the _default_ remote address
// to send messages and bundles to.
func NewConnection(lport int, raddr string) (*Connection, error) {
	var err error
	conn := &Connection{}
	if conn.laddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", lport)); err != nil {
		return nil, err
	}
	if raddr != "" {
		if conn.raddr, err = net.ResolveUDPAddr("udp", raddr); err != nil {
			return nil, err
		}
	}
	conn.errorCh = make(chan error)
	conn.readCh = make(chan readData, 20)
	return conn, nil
}

func (*Connection) ip(addr *net.UDPAddr) string {
	if addr == nil {
		return ""
	}
	return addr.IP.String()
}
func (c *Connection) RemoteAddress() string {
	return c.ip(c.raddr)
}
func (c *Connection) LocalAddress() string {
	return c.ip(c.laddr)
}

func (*Connection) port(addr *net.UDPAddr) int {
	if addr == nil {
		return -1
	}
	return addr.Port
}
func (c *Connection) RemotePort() int {
	return c.port(c.raddr)
}
func (c *Connection) LocalPort() int {
	return c.port(c.laddr)
}

func (c *Connection) SetLocalAddress(addr string) error {
	newAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	c.laddr = newAddr
	return nil
}

func (c *Connection) SetRemoteAddress(addr string) error {
	if addr == "" {
		c.raddr = nil
		return nil
	}
	newAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	c.raddr = newAddr
	return nil
}

// Send sends an OSC Bundle or an OSC Message.
func (c *Connection) Send(packet Packet) error {
	if c.conn == nil {
		if err := c.Open(); err != nil {
			return err
		}
	}

	data, err := packet.MarshalBinary()
	if err != nil {
		return err
	}

	_, err = c.conn.WriteToUDP(data, c.raddr)
	return err
}

func (c *Connection) Open() error {
	if c.conn != nil {
		return fmt.Errorf("connection already opened")
	}
	if c.Dispatcher == nil {
		c.Dispatcher = NewStandardDispatcher()
	}
	var err error
	c.conn, err = net.ListenUDP("udp", c.laddr)
	if err != nil {
		return err
	}
	return nil
}

func (c *Connection) Close() error {
	if c.doneCh != nil {
		close(c.doneCh)
	}
	if c.conn == nil {
		return fmt.Errorf("connection not open")
	}
	err := c.conn.Close()
	// If we get "use of closed network connection", it's not a problem because
	// closing the network connection is exactly what we wanted to do!
	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		return err
	}
	return nil
}

// Serve retrieves incoming OSC packets from the given connection and dispatches
// retrieved OSC packets. If something goes wrong an error is returned.
func (c *Connection) Serve() error {
	defer fmt.Println("exiting Serve())")
	c.doneCh = make(chan bool)
	go c.readFromConnection()
	var tempDelay time.Duration
	for {
		//	case msg, addr, err := c.readFromConnection():
		select {
		case d := <-c.readCh:
			go c.Dispatcher.Dispatch(d.msg, d.addr)
		case err := <-c.errorCh:
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
			return err
		case _, ok := <-c.doneCh:
			if !ok {
				return nil
			}
		}
	}
}

type readData struct {
	msg  Packet
	addr net.Addr
}

// readFromConnection retrieves OSC packets.
func (c *Connection) readFromConnection() {
	defer fmt.Println("exiting needFromConnection()")
	for {
		if c.ReadTimeout != 0 {
			if err := c.conn.SetReadDeadline(time.Now().Add(c.ReadTimeout)); err != nil {
				c.errorCh <- err
				return
			}
		}

		data := make([]byte, 65535)
		n, addr, err := c.conn.ReadFrom(data)
		if err != nil {
			c.errorCh <- err
			return
		}

		p, err := readPacket(bufio.NewReader(bytes.NewBuffer(data[0:n])))
		if err != nil {
			c.errorCh <- err
			return
		}

		c.readCh <- readData{msg: p, addr: addr}
		select {
		case _, ok := <-c.doneCh:
			if !ok {
				return
			}
		default:
		}
	}
}
