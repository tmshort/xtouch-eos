// Package osc provides a package for sending and receiving OpenSoundControl
// messages. The package is implemented in pure Go.
package osc

import (
	"bytes"
	"encoding/binary"
	"time"
)

const (
	secondsFrom1900To1970  = 2208988800                      // Source: RFC 868
	nanosecondsPerFraction = float64(0.23283064365386962891) // 1e9/(2^32)
)

// Timetag represents an OSC Time Tag.
// An OSC Time Tag is defined as follows:
// Time tags are represented by a 64 bit fixed point number. The first 32 bits
// specify the number of seconds since midnight on January 1, 1900, and the
// last 32 bits specify fractional parts of a second to a precision of about
// 200 picoseconds. This is the representation used by Internet NTP timestamps.
type Timetag struct {
	timeTag  uint64 // The acutal time tag
	time     time.Time
	MinValue uint64 // Minimum value of an OSC Time Tag. Is always 1.
}

////
// Timetag
////

// NewTimetag returns a new OSC time tag object.
func NewTimetag(ts time.Time) *Timetag {
	return &Timetag{
		time:     ts,
		timeTag:  timeToTimetag(ts),
		MinValue: uint64(1)}
}

// NewTimetagFromTimetag creates a new Timetag from the given `timetag`.
func NewTimetagFromTimetag(timetag uint64) *Timetag {
	time := timetagToTime(timetag)
	return NewTimetag(time)
}

// Time returns the time.
func (t *Timetag) Time() time.Time {
	return t.time
}

// FractionalSecond returns the last 32 bits of the OSC time tag. Specifies the
// fractional part of a second.
func (t *Timetag) FractionalSecond() uint32 {
	return uint32(t.timeTag << 32)
}

// SecondsSinceEpoch returns the first 32 bits (the number of seconds since the
// midnight 1900) from the OSC time tag.
func (t *Timetag) SecondsSinceEpoch() uint32 {
	return uint32(t.timeTag >> 32)
}

// TimeTag returns the time tag value
func (t *Timetag) TimeTag() uint64 {
	return t.timeTag
}

// MarshalBinary converts the OSC time tag to a byte array.
func (t *Timetag) MarshalBinary() ([]byte, error) {
	data := new(bytes.Buffer)
	if err := binary.Write(data, binary.BigEndian, t.timeTag); err != nil {
		return []byte{}, err
	}
	return data.Bytes(), nil
}

// SetTime sets the value of the OSC time tag.
func (t *Timetag) SetTime(time time.Time) {
	t.time = time
	t.timeTag = timeToTimetag(time)
}

// ExpiresIn calculates the number of seconds until the current time is the
// same as the value of the time tag. It returns zero if the value of the
// time tag is in the past.
func (t *Timetag) ExpiresIn() time.Duration {
	// If the timetag is one the OSC method must be invoke immediately.
	// See https://ccrma.stanford.edu/groups/osc/spec-1_0.html#timetags.
	if t.timeTag <= 1 {
		return 0
	}

	tt := timetagToTime(t.timeTag)
	seconds := time.Until(tt)

	// Invoke the OSC method immediately if the timetag is before or equal to the current time
	if seconds <= 0 {
		return 0
	}

	return seconds
}

// timeToTimetag converts the given time to an OSC time tag.
//
// An OSC time tag is defined as follows:
// Time tags are represented by a 64 bit fixed point number. The first 32 bits
// specify the number of seconds since midnight on January 1, 1900, and the
// last 32 bits specify fractional parts of a second to a precision of about
// 200 picoseconds. This is the representation used by Internet NTP timestamps.
//
// The time tag value consisting of 63 zero bits followed by a one in the least
// significant bit is a special case meaning "immediately."
//
// See also https://ccrma.stanford.edu/groups/osc/spec-1_0.html#timetags.
func timeToTimetag(v time.Time) (timetag uint64) {
	if v.IsZero() {
		// Means "immediately". It cannot occur otherwise as timetag == 0 gets
		// converted to January 1, 1900 while time.Time{} means year 1 in Go.
		// Use the IsZero method to detect it.
		return 1
	}

	seconds := uint64(v.Unix() + secondsFrom1900To1970)
	secondFraction := float64(v.Nanosecond()) / nanosecondsPerFraction

	return (seconds << 32) + uint64(uint32(secondFraction))
}

// timetagToTime converts the given OSC timetag to a time object.
func timetagToTime(timetag uint64) (t time.Time) {
	// Special case when timetag is == 1 that means "immediately". In this case we return
	// the zero time instant.
	if timetag == 1 {
		return time.Time{}
	}

	seconds := int64(timetag>>32) - secondsFrom1900To1970
	nanoseconds := int64(nanosecondsPerFraction * float64(float64(timetag&(1<<32-1))))

	return time.Unix(
		seconds,
		nanoseconds,
	)
}
