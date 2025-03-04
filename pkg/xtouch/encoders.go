package xtouch

import (
	"fmt"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

const (
	modeSingle     = 0x00
	modeFill       = 0x20
	modeWide       = 0x30
	modeBalance    = 0x10
	modeContinuous = 0x40
	ledOffset      = 47
)

type Encoder struct {
	base    *XTouch
	index   byte
	mode    byte
	value   byte
	handler func(byte, byte, int8)
}

func (e *Encoder) Handler(f func(byte, byte, int8)) *Encoder {
	e.handler = f
	return e
}

func (e *Encoder) callHandler(v uint8) error {
	delta := int8(v)
	if delta > 64 {
		delta = -(delta - 64)
	}

	value := int8(e.value) + delta
	if value < 1 {
		value = 1
	} else if value > 11 {
		value = 11
	}
	if e.mode == modeWide && value > 6 {
		value = 6
	}
	if e.value == byte(value) {
		return nil
	}
	if e.mode != modeContinuous {
		e.Set(byte(value))
	}
	if e.handler != nil {
		e.handler(e.index, e.Get(), delta)
	}
	return nil
}

func (e *Encoder) updateLeds(value byte) error {

	if e.mode == modeContinuous {
		return nil
	}
	// +47 is the index of the LED controls of the encoders
	return e.base.send(midi.ControlChange(e.base.channel, e.index+ledOffset, value))
}

func (e *Encoder) setMode(m byte) *Encoder {
	if e.index == 9 && m != modeContinuous {
		return nil
	}
	e.mode = m
	return e
}
func (e *Encoder) ModeSingle() *Encoder {
	return e.setMode(modeSingle)
}

func (e *Encoder) ModeFill() *Encoder {
	return e.setMode(modeFill)
}

func (e *Encoder) ModeWide() *Encoder {
	return e.setMode(modeWide)
}

func (e *Encoder) ModeBalance() *Encoder {
	return e.setMode(modeBalance)
}

func (e *Encoder) ModeContinuous() *Encoder {
	return e.setMode(modeContinuous)
}

func (e *Encoder) Off() error {
	return e.Set(0)
}

func (e *Encoder) Set(value byte) error {
	if e.mode == modeContinuous {
		return fmt.Errorf("can't set value in continuous mode")
	}
	e.value = value & 0x0F
	if value > 11 {
		value = 11
	}
	if e.mode == modeWide && value > 6 {
		value = 6
	}
	return e.updateLeds(e.value | e.mode)
}

func (e *Encoder) Get() byte {
	return e.value
}
