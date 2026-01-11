package config

import (
	"strconv"
	"time"
)

type SecondsDuration time.Duration

func (d *SecondsDuration) String() string {
	return time.Duration(*d).String()
}

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
