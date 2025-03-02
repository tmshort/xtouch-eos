package main

import (
	"fmt"
	"time"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

const (
	midiPort = "X-Touch"
)

func main() {
	fmt.Println("Hello main")

	defer midi.CloseDriver()

	inPorts := midi.GetInPorts()
	outPorts := midi.GetOutPorts()

	fmt.Printf("inPorts: %+v\n", inPorts)
	fmt.Printf("outPorts: %+v\n", outPorts)

	in, err := midi.FindInPort(midiPort)
	if err != nil {
		fmt.Printf("error getting in %s: %v\n", midiPort, err)
	}
	out, err := midi.FindOutPort(midiPort)
	if err != nil {
		fmt.Printf("error getting out %s: %v\n", midiPort, err)
	}
	send, err := midi.SendTo(out)
	if err != nil {
		fmt.Printf("error sending: %v\n", err)
	}

	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		fmt.Printf("Message %v\n", msg)
	}, midi.UseSysEx())

	fmt.Printf("Listening...\n")
	if err != nil {
		fmt.Printf("error listening: %v\n", err)
	}
	disp := &Displays{
		Send: send,
	}

	disp.SetLCDRaw("Hello World", 0)
	disp.SetLCDPanel(4, "Hello", "World44444444")
	disp.SetAcrossAll("  E0S 3.14.0")
	time.Sleep(time.Second * 60)

	disp.SetAcrossAll("")
	stop()
}

type Displays struct {
	Send func(midi.Message) error
}

func (d *Displays) SetLCDRaw(text string, startpos byte) {
	command := []byte{0x00, 0x00, 0x66, 0x14} // SysEx without 0xF0
	command = append(command, 0x12)           // LCD
	command = append(command, startpos)

	data := []byte(text)
	max := int(112 - startpos)
	if len(text) > max {
		data = data[:max]
	}
	command = append(command, data...)

	err := d.Send(midi.SysEx(command))
	if err != nil {
		fmt.Printf("error sending: %v\n", err)
	}
}

func (d *Displays) SetLCDPanel(panel int8, textTop, textBottom string) {
	posTop := (panel - 1) * 7
	posBottom := posTop + 0x38

	textTop = fmt.Sprintf("%-7.7s", textTop)
	textBottom = fmt.Sprintf("%-7.7s", textBottom)
	d.SetLCDRaw(textTop, byte(posTop))
	d.SetLCDRaw(textBottom, byte(posBottom))
}

func (*Displays) translateTextToMcu(text string) ([]byte, error) {
	mcu := map[rune]byte{
		'A': 1, 'B': 2, 'C': 3, 'D': 4, 'E': 5, 'F': 6, 'G': 7, 'H': 8, 'I': 9, 'J': 10, 'K': 11, 'L': 12, 'M': 13, 'N': 14, 'O': 15, 'P': 16, 'Q': 17, 'R': 18, 'S': 19, 'T': 20, 'U': 21, 'V': 22, 'W': 23, 'X': 24, 'Y': 25, 'Z': 26,
		'a': 1, 'b': 2, 'c': 3, 'd': 4, 'e': 5, 'f': 6, 'g': 7, 'h': 8, 'i': 9, 'j': 10, 'k': 11, 'l': 12, 'm': 13, 'n': 14, 'o': 15, 'p': 16, 'q': 17, 'r': 18, 's': 19, 't': 20, 'u': 21, 'v': 22, 'w': 23, 'x': 24, 'y': 25, 'z': 26,
		'0': 48, '1': 49, '2': 50, '3': 51, '4': 52, '5': 53, '6': 54, '7': 55, '8': 56, '9': 57,
		' ': 0, '-': 45, '_': 46,
	}

	var data []byte
	for _, c := range text {
		if n, ok := mcu[c]; ok {
			data = append(data, n)
		} else if c == '.' && len(data) > 0 && data[len(data)-1] < 64 {
			data[len(data)-1] += 64
		} else {
			return nil, fmt.Errorf("not a valid rune %v", c)
		}
	}
	return data, nil
}

func (d *Displays) doSevenSegment(position, value []byte) {
	for i := range position {
		d.Send(midi.ControlChange(0, position[i], value[i]))
	}
}

func (d *Displays) SetAssignment(text string) error {
	data, err := d.translateTextToMcu(text)
	if err != nil {
		return err
	}
	data = append(data, []byte{0, 0}...)
	d.doSevenSegment([]byte{75, 74}, data)
	return nil
}

func (d *Displays) SetBars(text string) error {
	data, err := d.translateTextToMcu(text)
	if err != nil {
		return err
	}
	data = append(data, []byte{0, 0, 0}...)
	d.doSevenSegment([]byte{73, 72, 71}, data)
	return nil
}

func (d *Displays) SetBeats(text string) error {
	data, err := d.translateTextToMcu(text)
	if err != nil {
		return err
	}
	data = append(data, []byte{0, 0}...)
	d.doSevenSegment([]byte{70, 69}, data)
	return nil
}

func (d *Displays) SetSubdivision(text string) error {
	data, err := d.translateTextToMcu(text)
	if err != nil {
		return err
	}
	data = append(data, []byte{0, 0}...)
	d.doSevenSegment([]byte{68, 67}, data)
	return nil

}
func (d *Displays) SetTicks(text string) error {
	data, err := d.translateTextToMcu(text)
	if err != nil {
		return err
	}
	data = append(data, []byte{0, 0, 0}...)
	d.doSevenSegment([]byte{66, 65, 64}, data)
	return nil
}

func (d *Displays) SetAcrossAll(text string) error {
	data, err := d.translateTextToMcu(text)
	if err != nil {
		return err
	}
	data = append(data, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}...)
	d.doSevenSegment([]byte{75, 74, 73, 72, 71, 70, 69, 68, 67, 66, 65, 64}, data)
	return nil
}
