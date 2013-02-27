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

	tickStepMs := time.Duration(100)
	tkr := time.NewTicker(tickStepMs * time.Millisecond)
	defer tkr.Stop()

	var nextScript *srt.Script = &b[0]
	var currScript *srt.Script = nil
	var currMs time.Duration
	/* var viasMs int */
	for {
		<-tkr.C
		currMs += tickStepMs
		/* currMs += viasMs */
		if currMs < 0 {
			nextScript = &b[0]
			continue
		}

		fmt.Printf("\r%d\t", currMs)

		if currScript != nil && time.Duration(currScript.EndMs) <= currMs {
			currScript = nil
		}

		if currScript != nil {
			fmt.Print(currScript.Text)
			/* continue */
		} else {
			fmt.Print("                           ")
		}

		if nextScript == nil {
			i := sort.Search(len(b), func(i int) bool {
				return time.Duration(b[i].StartMs) >= currMs
			})

			if i < len(b) {
				nextScript = &b[i]
			}
		}

		if nextScript != nil && time.Duration(nextScript.StartMs) <= currMs {
			currScript = nextScript
			nextScript = nil
		}

		if currScript == nil && nextScript == nil {
			break
		}
	}
}
