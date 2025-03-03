package xtouch

import (
	"fmt"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

type LcdDisplay struct {
	base  *XTouch
	index byte
}

func (l LcdDisplay) displayLcdRaw(text string, startpos byte) {
	command := []byte{0x00, 0x00, 0x66, 0x14} // SysEx without 0xF0
	command = append(command, 0x12)           // LCD
	command = append(command, startpos)

	data := []byte(text)
	max := int(112 - startpos)
	if len(text) > max {
		data = data[:max]
	}
	command = append(command, data...)

	err := l.base.send(midi.SysEx(command))
	if err != nil {
		fmt.Printf("error sending: %v\n", err)
	}
}

func (l LcdDisplay) SetPanel(textTop, textBottom string) {
	posTop := (l.index - 1) * 7
	posBottom := posTop + 0x38

	textTop = fmt.Sprintf("%-7.7s", textTop)
	textBottom = fmt.Sprintf("%-7.7s", textBottom)
	l.displayLcdRaw(textTop, byte(posTop))
	l.displayLcdRaw(textBottom, byte(posBottom))
}

func (l LcdDisplay) ClearPanel() {
	l.SetPanel("", "")
}
