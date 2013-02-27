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
	b := srt.UnmarshalFile(os.Args[1])

	screen := NewSdlContext(640, 480)
	defer screen.Release()

	screen.Clear()

	tickDuration, _ := time.ParseDuration("100ms")
	tkr := time.NewTicker(tickDuration)
	defer tkr.Stop()

	var nextScript *srt.Script = &b[0]
	startTime := time.Now()
	/* var viasMs int */
	for {
		<-tkr.C
		currMs := time.Since(startTime)
		fmt.Println("currMs", currMs)

		if currMs < 0 {
			nextScript = &b[0]
			continue
		}

		if nextScript == nil {
			fmt.Println("searching next script...")
			i := sort.Search(len(b), func(i int) bool {
				return time.Duration(b[i].StartMs)*time.Millisecond >= currMs
			})

			if i >= len(b) {
				lastScript := b[len(b)-1]
				if time.Duration(lastScript.EndMs) < currMs {
					fmt.Println("book ended")
					break
				}
			}
			nextScript = &b[i]
		}

		if nextScript != nil && time.Duration(nextScript.StartMs) <= currMs {
			screen.DisplayScript(nextScript)
			nextScript = nil
		}

	}
}
