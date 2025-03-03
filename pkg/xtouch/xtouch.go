package xtouch

import (
	"fmt"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

const (
	MidiPort = "X-Touch"
)

type XTouch struct {
	LedDisplay LedDisplay
	send       func(midi.Message) error
	stop       func()
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
		fmt.Printf("error getting in %s: %v\n", name, err)
	}

	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		fmt.Printf("Message %v\n", msg)
		midisend(msg)
	}, midi.UseSysEx())
	if err != nil {
		fmt.Printf("error listening: %v\n", err)
	}

	return &XTouch{
		LedDisplay: LedDisplay{send: midisend},
		send:       midisend,
		stop:       stop,
	}, nil
}

func (x *XTouch) LcdDisplay(i byte) LcdDisplay {
	return LcdDisplay{send: x.send, index: i}
}

func (x *XTouch) Encoder(i byte) Encoder {
	return Encoder{send: x.send, index: i + 47}
}

func (x *XTouch) Stop() {
	x.stop()
}
