package main

import (
	"fmt"
	"os"
	"time"

	"github.com/tmshort/xtouch-eos/pkg/xtouch"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	fmt.Println("Hello main")

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

	fmt.Printf("Listening...\n")

	//disp.SetLCDRaw("Hello World", 0)
	disp.LcdDisplay(4).SetPanel("Hello", "World44444444")
	disp.LedDisplay.SetAll("  E0S 3.14.0")
	disp.Encoder(1).Single(4)
	disp.Encoder(2).Balance(4)
	disp.Encoder(3).Fill(4)
	disp.Encoder(4).Wide(4)
	time.Sleep(time.Second * 60)

	disp.LedDisplay.SetAll("")
	for i := 0; i < 150; i++ {
		disp.Led(byte(i)).Off()
	}
	disp.Stop()
}
