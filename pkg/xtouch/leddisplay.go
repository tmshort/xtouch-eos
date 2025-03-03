package xtouch

import (
	"fmt"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

type LedDisplay struct {
	base *XTouch
}

type Led struct {
	base  *XTouch
	index byte
}

func (l Led) On() {
	l.base.send(midi.NoteOn(l.base.channel, l.index, 127))
}

func (l Led) Off() {
	l.base.send(midi.NoteOn(l.base.channel, l.index, 0))
}

func (LedDisplay) translateTextToMcu(text string) ([]byte, error) {
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

func (l LedDisplay) displaySevenSegment(position, value []byte) {
	for i := range position {
		l.base.send(midi.ControlChange(l.base.channel, position[i], value[i]))
	}
}

func (l LedDisplay) SetAssignment(text string) error {
	data, err := l.translateTextToMcu(text)
	if err != nil {
		return err
	}
	data = append(data, []byte{0, 0}...)
	l.displaySevenSegment([]byte{75, 74}, data)
	return nil
}

func (l LedDisplay) SetBars(text string) error {
	data, err := l.translateTextToMcu(text)
	if err != nil {
		return err
	}
	data = append(data, []byte{0, 0, 0}...)
	l.displaySevenSegment([]byte{73, 72, 71}, data)
	return nil
}

func (l LedDisplay) SetBeats(text string) error {
	data, err := l.translateTextToMcu(text)
	if err != nil {
		return err
	}
	data = append(data, []byte{0, 0}...)
	l.displaySevenSegment([]byte{70, 69}, data)
	return nil
}

func (l LedDisplay) SetSubdivision(text string) error {
	data, err := l.translateTextToMcu(text)
	if err != nil {
		return err
	}
	data = append(data, []byte{0, 0}...)
	l.displaySevenSegment([]byte{68, 67}, data)
	return nil
}

func (l LedDisplay) SetTicks(text string) error {
	data, err := l.translateTextToMcu(text)
	if err != nil {
		return err
	}
	data = append(data, []byte{0, 0, 0}...)
	l.displaySevenSegment([]byte{66, 65, 64}, data)
	return nil
}

func (l LedDisplay) SetAll(text string) error {
	data, err := l.translateTextToMcu(text)
	if err != nil {
		return err
	}
	data = append(data, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}...)
	l.displaySevenSegment([]byte{75, 74, 73, 72, 71, 70, 69, 68, 67, 66, 65, 64}, data)
	return nil
}
