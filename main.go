package main

import (
	"./srt"
	"fmt"
	"os"
	"sort"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:", os.Args[0], "[srt file]")
		return
	}
	b := srt.ReadSrtFile(os.Args[1])

	screen := NewSdlContext(640, 480)
	defer screen.Release()

	screen.Clear()

	tickDuration, _ := time.ParseDuration("10ms")
	tkr := time.NewTicker(tickDuration)
	defer tkr.Stop()

	var vias time.Duration
	var nextScript *srt.Script
	startTime := time.Now()
	for {
		<-tkr.C
		currMs := time.Since(startTime)
		currMs += vias

		if currMs < 0 {
			nextScript = &b[0]
			continue
		}

		if nextScript == nil {
			i := sort.Search(len(b), func(i int) bool {
				return b[i].Start >= currMs
			})

			if i >= len(b) {
				lastScript := b[len(b)-1]
				if lastScript.End < currMs {
					fmt.Println("book ended")
					break
				}
			}
			nextScript = &b[i]
		}

		if nextScript != nil && nextScript.Start <= currMs {
			screen.DisplayScript(nextScript)
			nextScript = nil
		}
	}
}
