package xtouch

import (
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

const (
	pitchLimit = 8192
)

type Fader struct {
	index   byte // 0-based, even though the public index is 1-based
	limit   int32
	value   int32 // kept as int32 for now, holds the absolute value 0 ~ 16383
	handler func(byte, uint16)
	base    *XTouch
}

func (f *Fader) Handler(h func(byte, uint16)) *Fader {
	f.handler = h
	return f
}

func (f *Fader) callHandler(abs uint16) error {
	f.setAbsolute(abs)
	if f.handler != nil {
		f.handler(f.index+1, f.Get())
	}
	return nil
}
func (f *Fader) Set(level uint16) error {
	f.value = (int32(level) * pitchLimit * 2) / f.limit
	return f.base.send(midi.Pitchbend(f.index, int16(f.value-pitchLimit)))
}

func (f *Fader) setAbsolute(level uint16) error {
	f.value = int32(level)
	return f.base.send(midi.Pitchbend(f.index, int16(f.value-pitchLimit)))
}

func (f *Fader) SetLimit(limit uint16) {
	f.limit = int32(limit)
}

func (f *Fader) Get() uint16 {
	return uint16((f.limit * f.value) / (pitchLimit * 2))
}
