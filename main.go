package main

import (
	"./subtitle"
	"fmt"
	"os"
	"sort"
	"time"
)

var viasC chan time.Duration

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:", os.Args[0], "[srt file]")
		return
	}
	book := subtitle.ReadSrtFile(os.Args[1])

	screen := NewSdlContext(640, 480)
	/* defer screen.Release() */

	screen.Clear()

	tickDuration, _ := time.ParseDuration("100ms")
	tkr := time.NewTicker(tickDuration)
	defer tkr.Stop()

	var vias time.Duration
	viasC = make(chan time.Duration)

	var nextScript *subtitle.Script
	startTime := time.Now()

CHAN_LOOP:
	for {
		select {
		case viasAdd := <-viasC:
			fmt.Println("vias=", viasAdd)
			vias += viasAdd
		case <-tkr.C:
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
						fmt.Println("book ended")
						break CHAN_LOOP
					}
				}
			}

			if nextScript != nil && nextScript.Start <= currMs {
				screen.DisplayScript(nextScript)
				nextScript = nil
			}
		}
	}
}
