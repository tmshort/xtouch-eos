package xtouch

import (
	"fmt"
	"os"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

const (
	MidiPort = "X-Touch"
)

var (
	nameToNote map[string]byte
	noteToName map[byte]string
)

type XTouch struct {
	LedDisplay LedDisplay
	channel    byte
	send       func(midi.Message) error
	stop       func()
}

func init() {
	nameToNote = map[string]byte{
		"TRACK":            40,
		"PAN/SURROUND":     42,
		"EQ":               44,
		"SEND":             41,
		"PLUG-IN":          43,
		"INST":             45,
		"GLOBAL VIEW":      51,
		"MIDI TRACKS":      62,
		"INPUTS":           63,
		"AUDIO TRACKS":     64,
		"AUDIO INST":       65,
		"AUX":              66,
		"BUSES":            67,
		"OUTPUTS":          68,
		"USER":             69,
		"F1":               54,
		"F2":               55,
		"F3":               56,
		"F4":               57,
		"F5":               58,
		"F6":               59,
		"F7":               60,
		"F8":               61,
		"SHIFT":            70,
		"OPTION":           71,
		"CONTROL":          72,
		"ALT":              73,
		"READ/OFF":         74,
		"WRITE":            75,
		"TRIM":             76,
		"TOUCH":            77,
		"LATCH":            78,
		"GROUP":            79,
		"SAVE":             80,
		"UNDO":             81,
		"CANCEL":           82,
		"ENTER":            83,
		"MARKER":           84,
		"NUDGE":            85,
		"CYCLE":            86,
		"DROP":             87,
		"REPLACE":          88,
		"CLICK":            89,
		"SOLO":             90,
		"REWIND":           91,
		"FAST-FORWARD":     92,
		"STOP":             93,
		"PLAY":             94,
		"RECORD":           95,
		"FADER BANK/LEFT":  46,
		"FADER BANK/RIGHT": 47,
		"CHANNEL/LEFT":     48,
		"CHANNEL/RIGHT":    49,
		"SCRUB":            101,
		"UP":               96,
		"LEFT":             98,
		"CENTER":           100,
		"RIGHT":            99,
		"DOWN":             97,
		"FLIP":             50,
		"NAME/VALUE":       52,
		"BEATS":            53,
		"REC/1":            0,
		"REC/2":            1,
		"REC/3":            2,
		"REC/4":            3,
		"REC/5":            4,
		"REC/6":            5,
		"REC/7":            6,
		"REC/8":            7,
		"SOLO/1":           8,
		"SOLO/2":           9,
		"SOLO/3":           10,
		"SOLO/4":           11,
		"SOLO/5":           12,
		"SOLO/6":           13,
		"SOLO/7":           14,
		"SOLO/8":           15,
		"MUTE/1":           16,
		"MUTE/2":           17,
		"MUTE/3":           18,
		"MUTE/4":           19,
		"MUTE/5":           20,
		"MUTE/6":           21,
		"MUTE/7":           22,
		"MUTE/8":           23,
		"SELECT/1":         24,
		"SELECT/2":         25,
		"SELECT/3":         26,
		"SELECT/4":         27,
		"SELECT/5":         28,
		"SELECT/6":         29,
		"SELECT/7":         30,
		"SELECT/8":         31,
		"ENCODER/1":        32,
		"ENCODER/2":        33,
		"ENCODER/3":        34,
		"ENCODER/4":        35,
		"ENCODER/5":        36,
		"ENCODER/6":        37,
		"ENCODER/7":        38,
		"ENCODER/8":        39,
		"SMPTE/LED":        113,
		"BEATS/LED":        114,
		"SOLO/LED":         115,
	}
	// This tests for duplicates
	noteToName = map[byte]string{}
	for name, note := range nameToNote {
		if n, ok := noteToName[note]; ok {
			fmt.Printf("noteToName: found name %v for note %v\n", n, note)
			os.Exit(1)
		}
		noteToName[note] = name
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

	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		fmt.Printf("Message %v\n", msg)
		midisend(msg)
	}, midi.UseSysEx())
	if err != nil {
		return nil, err
	}

	x := &XTouch{
		send: midisend,
		stop: stop,
	}

	x.LedDisplay = LedDisplay{base: x}
	return x, nil
}

func (x *XTouch) LcdDisplay(i byte) LcdDisplay {
	return LcdDisplay{base: x, index: i}
}

func (x *XTouch) Encoder(i byte) Encoder {
	return Encoder{base: x, index: i + 47}
}

func (x *XTouch) Led(i byte) Led {
	return Led{base: x, index: i}
}

func (x *XTouch) Button(name string) Led {
	if note, ok := nameToNote[name]; ok {
		return Led{base: x, index: note}
	}
	return Led{base: x, index: 0}

}

func (x *XTouch) Stop() {
	x.stop()
}
