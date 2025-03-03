package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tmshort/xtouch-eos/pkg/xtouch"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	defer midi.CloseDriver()

	inPorts := midi.GetInPorts()
	outPorts := midi.GetOutPorts()

	fmt.Printf("inPorts: %+v\n", inPorts)
	fmt.Printf("outPorts: %+v\n", outPorts)

	disp, err := xtouch.NewXTouch()
	if err != nil {
		fmt.Printf("error creating XTouch: %v\n", err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 20)
	go signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	timeChan := time.NewTimer(time.Minute).C

	fmt.Printf("Listening...\n")

	//disp.SetLCDRaw("Hello World", 0)
	disp.LcdDisplay(4).SetPanel("Hello", "World44444444")
	disp.LedDisplay.SetAll("  E0S 3.14.0")

	encoderHandler := func(i byte, v byte, d int8) {
		fmt.Printf("index=%v value=%v delta=%v\n", i, v, d)
	}

	disp.Encoder(1).ModeSingle().Handler(encoderHandler).Set(4)
	disp.Encoder(2).ModeBalance().Handler(encoderHandler).Set(4)
	disp.Encoder(3).ModeFill().Handler(encoderHandler).Set(4)
	disp.Encoder(4).ModeWide().Handler(encoderHandler).Set(4)
	disp.Encoder(5).ModeContinuous().Handler(encoderHandler)
	disp.Encoder(9).ModeWide().Handler(encoderHandler)

	handler := func(name string, note byte, value bool) {
		fmt.Printf("Button %v (%v) = %v\n", name, note, value)
	}

	disp.Button("MARKER").PressBehavior().Handler(handler)
	disp.Button("NUDGE").ToggleBehavior().Handler(handler)

	disp.Fader(2).Handler(func(f byte, v uint16) {
		fmt.Printf("fader %v at %v\n", f, v)
	})

	func() {
		select {
		case <-timeChan:
			return
		case <-sigChan:
			return
		}
	}()

	fmt.Printf("\nCleaning up...\n")
	disp.Stop()
}
