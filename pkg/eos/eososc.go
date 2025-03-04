package eos

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hypebeast/go-osc/osc"
)

const (
	defaultPort = 8765
)

type Eos struct {
	client     *osc.Client
	server     *osc.Server
	dispatcher *osc.StandardDispatcher
}

func NewEos(txaddr, rxaddr string) (*Eos, error) {
	var host string
	var port int
	var err error

	args := strings.Split(txaddr, ":")
	switch {
	case len(args) == 2:
		host = args[0]
		port, err = strconv.Atoi(args[1])
		if err != nil {
			return nil, err
		}
	case len(args) == 1:
		host = args[0]
		port = defaultPort
	default:
		return nil, fmt.Errorf("invalid txaddr: %v", txaddr)
	}
	client := osc.NewClient(host, port)

	args = strings.Split(rxaddr, ":")
	switch {
	case len(args) == 2:
		host = args[0]
		port, err = strconv.Atoi(args[1])
		if err != nil {
			return nil, err
		}
	case len(args) == 1:
		host = args[0]
		port = defaultPort
	default:
		return nil, fmt.Errorf("invalid rxaddr: %v", rxaddr)
	}
	dispatcher := osc.NewStandardDispatcher()
	server := &osc.Server{
		Addr:       fmt.Sprintf("%v:%v", host, port),
		Dispatcher: dispatcher,
	}
	return &Eos{client: client, server: server}, nil
}

func (e *Eos) Handler(prefix string, handler func(msg *osc.Message)) error {
	return e.dispatcher.AddMsgHandler(prefix, handler)
}
