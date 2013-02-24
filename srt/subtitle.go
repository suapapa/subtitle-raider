package srt

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	/* "time" */
)

type Script struct {
	Idx     int
	StartMs int
	EndMs   int
	Text    string
}

type Book []Script

func Unmarshal(data []byte) Book {
	var book Book
	var script Script

	b := bytes.NewBuffer(data)

	const (
		STATE_IDX = iota
		STATE_TS
		STATE_SCRIPT
	)

	state := STATE_IDX
	for {
		line, err := b.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimRight(line, "\r\n")
		/* log.Printf("line = '%s'", line) */

		switch state {
		case STATE_IDX:
			/* log.Println("STATE_IDX") */
			_, err := fmt.Sscanln(line, &script.Idx)
			if err != nil {
				log.Fatalln("failed to parse index!", err)
			}
			state = STATE_TS

		case STATE_TS:
			/* log.Println("STATE_TS") */
			var sH, sM, sS, sMs int
			var eH, eM, eS, eMs int
			_, err := fmt.Sscanf(line,
				"%d:%d:%d,%d --> %d:%d:%d,%d",
				&sH, &sM, &sS, &sMs,
				&eH, &eM, &eS, &eMs)
			if err != nil {
				log.Fatalln("failed to parse timestamp!")
			}
			script.StartMs = sMs + sS*1000 + sM*60*1000 + sH*60*60*1000
			script.EndMs = eMs + eS*1000 + eM*60*1000 + eH*60*60*1000
			script.Text = ""
			/* log.Println("script = ", script) */
			state = STATE_SCRIPT

		case STATE_SCRIPT:
			/* log.Println("STATE_SCRIPT") */
			if line == "" {
				/* log.Println("script = ", script) */
				book = append(book, script)
				state = STATE_IDX
			} else {
				if script.Text != "" {
					script.Text += "\n"
				}
				script.Text += line
			}
		}

	}
	/* log.Println("book = ", book) */
	return book
}

func UnmarshalFile(filename string) Book {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalln("faile to read file, ", filename)
	}
	return Unmarshal(data)
}
