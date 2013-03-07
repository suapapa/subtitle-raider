// Copyright 2013, Homin Lee. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package subtitle

import (
	"regexp"
	"time"
)

// Collection of scripts
type Book []Script

// A script
type Script struct {
	Idx        int
	Start, End time.Duration
	Text       string
}

// How long the script should be shown
func (s *Script) Duration() time.Duration {
	return s.End - s.Start
}

// Script HTML markup from text
func (s *Script) TextWithoutMarkup() string {
	return reMakrup.ReplaceAllString(s.Text, "")
}

// Check the script with given timestamp
func (s *Script) CheckHit(ts time.Duration) HitStatus {
	switch {
	case ts < s.Start:
		return HS_EARLY
	case ts >= s.Start && s.End >= s.End:
		return HS_HIT
	case s.End < ts:
		return HS_LATE
	}
	return HS_INVALID
}

// HitStatus is type for timestamp check
type HitStatus uint8

const (
	HS_INVALID HitStatus = iota
	HS_EARLY             // Not yet
	HS_HIT               // Now
	HS_LATE              // Gone
)

var reMakrup = regexp.MustCompile("</?[^<>]+?>")
