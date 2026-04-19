package config

import (
	"strconv"
	"time"
)

// SecondsDuration is a time.Duration that supports parsing from seconds or duration strings.
type SecondsDuration time.Duration

// String returns the duration in a human-readable format.
func (d *SecondsDuration) String() string {
	return time.Duration(*d).String()
}

// Set parses a duration from either an integer number of seconds
// or a standard time duration string (e.g. "10s", "1m").
func (d *SecondsDuration) Set(s string) error {
	if sec, err := strconv.Atoi(s); err == nil {
		*d = SecondsDuration(time.Duration(sec) * time.Second)
		return nil
	}
	td, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = SecondsDuration(td)
	return nil
}
