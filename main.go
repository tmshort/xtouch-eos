package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	//"github.com/hypebeast/go-osc/osc"
	"github.com/tmshort/xtouch-eos/pkg/eos"
	"github.com/tmshort/xtouch-eos/pkg/osc"
	"github.com/tmshort/xtouch-eos/pkg/xtouch"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	defer midi.CloseDriver()

	xt, err := xtouch.NewXTouch()
	if err != nil {
		fmt.Printf("error creating XTouch: %v\n", err)
		os.Exit(1)
	}
	defer xt.Stop()

	sigChan := make(chan os.Signal, 20)
	go signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	timeChan := time.NewTimer(time.Minute).C

	fmt.Printf("Listening...\n")

	//xt.SetLCDRaw("Hello World", 0)
	xt.LcdDisplay(4).SetPanel("Hello", "World44444444")
	xt.LedDisplay.SetAll("  E0S 3.14.0")

	encoderHandler := func(i byte, v byte, d int8) {
		fmt.Printf("index=%v value=%v delta=%v\n", i, v, d)
	}

	xt.Encoder(1).ModeSingle().Handler(encoderHandler).Set(4)
	xt.Encoder(2).ModeBalance().Handler(encoderHandler).Set(4)
	xt.Encoder(3).ModeFill().Handler(encoderHandler).Set(4)
	xt.Encoder(4).ModeWide().Handler(encoderHandler).Set(4)
	xt.Encoder(5).ModeContinuous().Handler(encoderHandler)
	xt.Encoder(9).Handler(encoderHandler)

	buttonHandler := func(name string, note byte, value bool) {
		fmt.Printf("Button %v (%v) = %v\n", name, note, value)
	}

	xt.Button("MARKER").PressBehavior().Handler(buttonHandler)
	xt.Button("NUDGE").ToggleBehavior().Handler(buttonHandler)

	xt.Fader(2).Handler(func(f byte, v uint16) {
		fmt.Printf("fader %v at %v\n", f, v)
	})

	//e, err := eos.NewEos("0.0.0.0", "192.168.1.222")
	e, err := eos.NewEos("0.0.0.0", "127.0.0.1")
	if err != nil {
		fmt.Printf("error creating Eos: %v\n", err)
		os.Exit(1)
	}
	defer e.Close()

	fmt.Printf("listening for OSC\n")
	err = e.Handler("/eos/out/get/version", func(msg *osc.Message, addr net.Addr) {
		fmt.Printf("Received from %v: %+v\n", addr, msg)
		for i, a := range msg.Arguments {
			fmt.Printf("Arg[%d]=%v\n", i, a)
		}
		xt.LedDisplay.SetAll(fmt.Sprintf("  E0S %v", msg.Arguments[0]))
	})
	if err != nil {
		fmt.Printf("error adding handler: %v\n", err)
	}

	e.StartServer()

	fmt.Printf("sending OSC\n")
	err = e.SendMessage(osc.NewMessage("/eos/get/version"))
	if err != nil {
		fmt.Printf("error sending message: %v\n", err)
	}

	func() {
		select {
		case <-timeChan:
			return
		case <-sigChan:
			return
		}
	}()

	fmt.Printf("\nCleaning up...\n")
}
