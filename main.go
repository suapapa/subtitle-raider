// Copyright 2013, Homin Lee. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"./subtitle"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	tsViasC chan time.Duration
	navC    chan int
	quitC   chan bool
)

func init() {
	tsViasC = make(chan time.Duration)
	navC = make(chan int)
	quitC = make(chan bool)
}

func main() {
	screen := NewScreen(opt.scrnW, opt.scrnH)

	go eventLoop(screen)
	go graphicLoop(screen)

	tkr := time.NewTicker(time.Millisecond * 100)
	debugTkr := time.NewTicker(time.Second / 10)
	if opt.debugScrn == false {
		debugTkr.Stop()
	}

	startTime := time.Now()

	var tsVias time.Duration
	var tsClear time.Duration
	var paused bool
	var currScriptIdx int

	var nextScript *subtitle.Script
	book := subtitle.ReadSrtFile(flags[0])

	if opt.startIdx > 0 && opt.startIdx < len(book) {
		go func() {
			navC <- opt.startIdx
		}()
	}
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
			screen.DisplayScript(currScript)
			if paused == false {
				startTime = time.Now()
				tsVias = currScript.Start
				tsClear = currScript.End
			} else {
				tsClear = 0
			}

		case v := <-tsViasC:
			tsVias += v

		case <-debugTkr.C:
			if nextScript == nil {
				continue CHAN_LOOP
			}
			tsCurr := time.Since(startTime)
			debugStr := fmt.Sprintf("%d:%s(%s...%s) ClearTs=%s Ts=%s",
				nextScript.Idx, nextScript.Text,
				nextScript.Start, nextScript.End,
				tsClear, tsCurr+tsVias)
			screen.displayDebug(debugStr)
			/* screen.displayDebug("debug display") */

		case <-tkr.C:
			if paused {
				continue
			}

			tsCurr := time.Since(startTime)
			tsCurr += tsVias
			if tsCurr < 0 {
				nextScript = &book[0]
				continue
			}

			if nextScript == nil {
				nextScript = book.Find(tsCurr)
			}

			if nextScript != nil {
				if hs := nextScript.CheckHit(tsCurr); hs == subtitle.SCR_HIT {
					screen.DisplayScript(nextScript)
					tsClear = nextScript.End
					currScriptIdx = nextScript.Idx - 1
					nextScript = nil
				}
			}

			if tsClear != 0 && tsClear <= tsCurr {
				screen.Clear()
				tsClear = 0
			}

		case <-quitC:
			screen.Release()
			log.Println("Bye Bye")
			os.Exit(0)

		}
	}
}
