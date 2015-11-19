// Copyright 2013, Homin Lee. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/suapapa/go_subtitle"
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

	scriptFileName := flags[0]
	f, err := os.Open(scriptFileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var book subtitle.Book
	switch {
	case strings.HasSuffix(scriptFileName, ".srt"):
		book, err = subtitle.ReadSrt(f)
		if err != nil {
			panic(err)
		}
	case strings.HasSuffix(scriptFileName, ".smi"):
		book, err = subtitle.ReadSmi(f)
		if err != nil {
			panic(err)
		}
	}

	if opt.startIdx > 0 && opt.startIdx < len(book) {
		go func() {
			navC <- opt.startIdx
		}()
	}

	var tsVias time.Duration
	var tsClear time.Duration
	var paused bool
	var nextScript *subtitle.Script

	startTime := time.Now()
	currScriptIdx := -1

CHAN_LOOP:
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
		tsCurr := time.Since(startTime) + tsVias
		debugStr := fmt.Sprintf("%s ClearTs=%s Ts=%s",
			nextScript, tsClear, tsCurr)
		screen.displayDebug(debugStr)

	case <-tkr.C:
		if paused {
			break
		}

		tsCurr := time.Since(startTime) + tsVias
		if tsCurr < 0 {
			nextScript = &book[0]
			break
		}

		if nextScript == nil {
			nextScript = book.Find(tsCurr)
		}

		if nextScript != nil {
			if subtitle.SCR_HIT == nextScript.CheckHit(tsCurr) {
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
	goto CHAN_LOOP
}
