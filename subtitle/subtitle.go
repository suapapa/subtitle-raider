package subtitle

import (
	"time"
)

type Script struct {
	Idx int
	Start, End time.Duration
	Text string
}

func (s *Script) Duration() time.Duration {
	return s.End - s.Start
}

type Book []Script
