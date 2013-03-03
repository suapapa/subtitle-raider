// Copyright 2013, Homin Lee. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package subtitle

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

var reFont = regexp.MustCompile("<font.*?>(.+)</font>")

func ReadSami(data []byte) Book {
	var book Book
	var script Script

	const (
		SAMI_SYNC = iota
		SAMI_CLASS
		SAMI_TEXT
	)

	state = SAMI_SYNC

	book = append(book, script)
	/* log.Println("book = ", book) */
	return book
}

func ReadSamiFile(filename string) Book {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalln("faile to read file, ", filename)
	}
	return ReadSami(data)
}
