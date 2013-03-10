// Copyright 2013, Homin Lee. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package subtitle

import (
	"fmt"
	"regexp"
	"sort"
	"time"
)

// Collection of scripts
type Book []Script

// Find a script on given timestamp.
// If not hit, it returns next script.
// So, caller should re-check the script is hit on the timestamp.
func (b Book) Find(ts time.Duration) *Script {
	si := sort.Search(len(b), func(i int) bool {
		return b[i].Start >= ts
	})

	if si >= len(b) {
		return nil
	}

	return &b[si]
}

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
		return SCR_EARLY
	case ts >= s.Start && s.End >= s.End:
		return SCR_HIT
	case s.End < ts:
		return SCR_LATE
	}
	return SCR_INVALID
}

func (s *Script) String() string {
	return fmt.Sprintf("%d:%s(%s-%s)", s.Idx, s.Text, s.Start, s.End)
}

// HitStatus is type for timestamp check
type HitStatus uint8

const (
	SCR_INVALID HitStatus = iota
	SCR_EARLY             // Not yet
	SCR_HIT               // Now
	SCR_LATE              // Gone
)

var reMakrup = regexp.MustCompile("</?[^<>]+?>")
