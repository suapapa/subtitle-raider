// Copyright 2013, Homin Lee. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package subtitle

import (
	"regexp"
	"time"
)

var reMakrup = regexp.MustCompile("</?[^<>]+?>")

type Script struct {
	Idx        int
	Start, End time.Duration
	Text       string
}

func (s *Script) Duration() time.Duration {
	return s.End - s.Start
}

func (s *Script) TextWithoutMarkup() string {
	return reMakrup.ReplaceAllString(s.Text, "")
}

type Book []Script
