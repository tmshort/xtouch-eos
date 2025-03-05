package eos

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/tmshort/xtouch-eos/pkg/osc"
	//"github.com/hypebeast/go-osc/osc"
)

const (
	defaultRemotePort = 53000 // send to this // 4703 // 8000
	defaultLocalPort  = 53001 // receive on this // 4704 // 8001
)

type Eos struct {
	conn       *osc.Connection
	dispatcher *osc.StandardDispatcher
}

func NewEos(laddr, raddr string) (*Eos, error) {
	var port int
	var err error

	args := strings.Split(raddr, ":")
	switch {
	case len(args) == 2:
		if _, err = strconv.Atoi(args[1]); err != nil {
			return nil, err
		}
	case len(args) == 1:
		raddr = fmt.Sprintf("%s:%d", args[0], defaultRemotePort)
	default:
		return nil, fmt.Errorf("invalid raddr: %v", raddr)
	}

	args = strings.Split(laddr, ":")
	switch {
	case len(args) == 2:
		if port, err = strconv.Atoi(args[1]); err != nil {
			return nil, err
		}
	case len(args) == 1:
		port = defaultLocalPort
	default:
		return nil, fmt.Errorf("invalid laddr: %v", laddr)
	}
	dispatcher := osc.NewStandardDispatcher()
	conn, err := osc.NewConnection(port, raddr)
	if err != nil {
		return nil, err
	}
	conn.Dispatcher = dispatcher

	return &Eos{conn: conn, dispatcher: dispatcher}, nil
}

func (e *Eos) Handler(prefix string, handler func(msg *osc.Message, addr net.Addr)) error {
	return e.dispatcher.AddMsgHandler(prefix, handler)
}

func (e *Eos) Close() error {
	return e.conn.Close()
}

func (e *Eos) StartServer() error {
	if err := e.conn.Open(); err != nil {
		return err
	}
	go e.conn.Serve()
	return nil
}

func (e *Eos) SendMessage(msg *osc.Message) error {
	return e.conn.Send(msg)
}
