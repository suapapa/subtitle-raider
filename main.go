// Copyright 2013, Homin Lee. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"./subtitle"
	"fmt"
	"os"
	"sort"
	"time"
)

var (
	viasC chan time.Duration
	navC  chan int
	quitC chan bool
)

func init() {
	viasC = make(chan time.Duration)
	navC = make(chan int)
	quitC = make(chan bool)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:", os.Args[0], "[srt file]")
		return
	}
	book := subtitle.ReadSrtFile(os.Args[1])

	// XXX: fix to get size form argument
	screen := NewSdlContext(1024, 480)
	/* defer screen.Release() */

	screen.Clear()

	tickDuration, _ := time.ParseDuration("100ms")
	tkr := time.NewTicker(tickDuration)
	/* defer tkr.Stop() */

	debugTkr := time.NewTicker(time.Second / 15)

	var nextScript *subtitle.Script
	startTime := time.Now()

	var vias time.Duration
	var paused bool
	var currScriptIdx int
CHAN_LOOP:
	for {
		select {
		case nav := <-navC:
			currScriptIdx += nav
			if currScriptIdx < 0 {
				currScriptIdx = 0
			} else if currScriptIdx >= len(book) {
				currScriptIdx = len(book) - 1
			}
			if nav == 0 {
				paused = !paused
			}

			currScript := &book[currScriptIdx]
			nextScript = nil
			if paused == false {
				startTime = time.Now()
				vias = -currScript.Start
			}
			screen.DisplayScript(currScript)

		case viasAdd := <-viasC:
			vias += viasAdd

		case <-debugTkr.C:
			currMs := time.Since(startTime)
			debugStr := fmt.Sprintf("currMS=%s, vias=%s, tunedTs=%s nextScript=%s",
				currMs, vias, currMs+vias, nextScript)
			screen.displayDebug(debugStr)

		case <-tkr.C:
			// XXX: pause not working?
			if paused {
				continue
			}
			currMs := time.Since(startTime)
			currMs += vias

			if currMs < 0 {
				nextScript = &book[0]
				continue
			}

			if nextScript == nil {
				i := sort.Search(len(book), func(i int) bool {
					return book[i].Start >= currMs
				})

				if i < len(book) {
					nextScript = &book[i]
				} else {
					lastScript := book[len(book)-1]
					if lastScript.End < currMs {
						break CHAN_LOOP
					}
				}
			}

			if nextScript != nil && nextScript.Start <= currMs {
				screen.DisplayScript(nextScript)
				currScriptIdx = nextScript.Idx - 1
				nextScript = nil
			}
		case <-quitC:
			break CHAN_LOOP

		}
	}
}
