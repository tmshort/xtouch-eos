package xtouch

import (
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

type Encoder struct {
	send  func(midi.Message) error
	index byte
}

func (e Encoder) Off() {
	e.send(midi.ControlChange(0, e.index, 0))
}

func (e Encoder) Single(value byte) {
	value &= 0x0F
	value |= 0x00
	e.send(midi.ControlChange(0, e.index, value))
}

func (e Encoder) Fill(value byte) {
	value &= 0x0F
	value |= 0x20
	e.send(midi.ControlChange(0, e.index, value))
}

func (e Encoder) Wide(value byte) {
	value &= 0x0F
	value |= 0x30
	e.send(midi.ControlChange(0, e.index, value))
}

func (e Encoder) Balance(value byte) {
	value &= 0x0F
	value |= 0x10
	e.send(midi.ControlChange(0, e.index, value))
}
