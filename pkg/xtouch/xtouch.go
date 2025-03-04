package xtouch

import (
	"fmt"
	"os"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

const (
	MidiPort = "X-Touch INT"
)

type XTouch struct {
	LedDisplay   LedDisplay
	channel      byte
	send         func(midi.Message) error
	stop         func()
	nameToNote   map[string]byte
	noteToButton map[byte]*Button
	faders       map[byte]*Fader
	encoders     map[byte]*Encoder
}

func (x *XTouch) initButtons() {
	x.nameToNote = buttonNameToNote
	// This tests for duplicates
	x.noteToButton = map[byte]*Button{}
	for name, note := range x.nameToNote {
		if n, ok := x.noteToButton[note]; ok {
			fmt.Printf("noteToName: found name %v for note %v\n", n, note)
			os.Exit(1)
		}
		x.noteToButton[note] = &Button{
			note:     note,
			name:     name,
			behavior: defaultButtonBehavior,
			base:     x,
		}
	}
}

func NewXTouch() (*XTouch, error) {
	return NewXTouchByName(MidiPort)
}

func NewXTouchByName(name string) (*XTouch, error) {
	out, err := midi.FindOutPort(name)
	if err != nil {
		return nil, err
	}
	midisend, err := midi.SendTo(out)
	if err != nil {
		return nil, err
	}

	in, err := midi.FindInPort(name)
	if err != nil {
		return nil, err
	}

	x := &XTouch{
		send: midisend,
	}
	x.initButtons()
	x.faders = map[byte]*Fader{}
	x.encoders = map[byte]*Encoder{}
	for i := 1; i <= 9; i++ {
		x.faders[byte(i)] = &Fader{index: byte(i) - 1, base: x, limit: 101}
		x.encoders[byte(i)] = &Encoder{index: byte(i), base: x, mode: modeContinuous}
	}

	x.stop, err = midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var ch, key, v uint8
		var u16 uint16
		var err error
		switch {
		case msg.GetNoteOn(&ch, &key, &v):
			err = x.ButtonByNote(key).callBehavior(msg)
		case msg.GetPitchBend(&ch, nil, &u16):
			err = x.Fader(ch + 1).callHandler(u16)
		case msg.GetControlChange(&ch, &key, &v):
			err = x.encoderFromController(key).callHandler(v)
		default:
			fmt.Printf("Message %v\n", msg)
			err = x.send(msg)
		}
		if err != nil {
			fmt.Printf("error handling msg: %v\n", msg)
		}
	}, midi.UseSysEx())
	if err != nil {
		return nil, err
	}

	x.LedDisplay = LedDisplay{base: x}
	return x, nil
}

func (x *XTouch) LcdDisplay(i byte) LcdDisplay {
	return LcdDisplay{base: x, index: i}
}

func (x *XTouch) encoderFromController(i byte) *Encoder {
	if 16 <= i && i <= 23 {
		return x.encoders[i-15]
	}
	if i == 60 {
		return x.encoders[9]
	}
	return nil
}
func (x *XTouch) Encoder(i byte) *Encoder {
	return x.encoders[i]
}

func (x *XTouch) Led(i byte) Led {
	return Led{base: x, index: i}
}

func (x *XTouch) ButtonByNote(note byte) *Button {
	return x.noteToButton[note]
}

func (x *XTouch) Button(name string) *Button {
	if note, ok := x.nameToNote[name]; ok {
		return x.noteToButton[note]
	}
	return nil
}

func (x *XTouch) Fader(index byte) *Fader {
	return x.faders[index]
}

func (x *XTouch) Stop() {
	x.LedDisplay.SetAll("")
	for i := 1; i <= 8; i++ {
		x.LcdDisplay(byte(i)).ClearPanel()
		x.Encoder(byte(i)).Off()
		x.Fader(byte(i)).setAbsolute(0)
	}
	x.Fader(9).setAbsolute(0)
	for _, b := range x.noteToButton {
		b.Off()
	}
	x.stop()
}
