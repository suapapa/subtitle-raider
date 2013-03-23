// Copyright 2013, Homin Lee. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	DFLT_FONT_PATH = "/usr/share/fonts/truetype/nanum/NanumGothicBold.ttf"
	DFLT_FONT_SIZE = 90
	DFLT_FPS       = 15
)

var (
	opt   Options
	flags []string
)

func init() {
	opt, flags = parseFlags()
}

type Options struct {
	fontSize     int
	fontPath     string
	startIdx     int
	scrnW, scrnH int
	fullscreen   bool
	fps          int
	alignCenter  bool
	debugScrn    bool

	showText string
}

func setupFlags(opts *Options) *flag.FlagSet {
	prgName := os.Args[0]
	fs := flag.NewFlagSet(prgName, flag.ExitOnError)
	fs.IntVar(&opts.fontSize, "fs", DFLT_FONT_SIZE, "font size")
	fs.StringVar(&opts.fontPath, "fp", DFLT_FONT_PATH, "font path")
	fs.IntVar(&opts.fps, "fps", DFLT_FPS, "fps")
	fs.IntVar(&opts.startIdx, "s", 0, "set first scipt idx")
	fs.IntVar(&opts.scrnW, "w", 1024, "screen width")
	fs.IntVar(&opts.scrnH, "h", 480, "screen height")
	fs.BoolVar(&opts.fullscreen, "f", false, "fullscreen")
	fs.BoolVar(&opts.alignCenter, "c", false, "center align")
	fs.BoolVar(&opts.debugScrn, "d", false, "show debug message on screen")

	fs.Usage = func() {
		fmt.Printf("Usage: %s [options] subtitle_file\n", prgName)
		fs.PrintDefaults()
	}

	return fs
}

func verifyFlags(opts *Options, fs *flag.FlagSet) {

	if len(fs.Args()) == 0 && opts.showText == "" {
		fs.Usage()
		os.Exit(1)
	}
}

func parseFlags() (Options, []string) {
	var opts Options
	fs := setupFlags(&opts)
	fs.Parse(os.Args[1:])
	verifyFlags(&opts, fs)
	return opts, fs.Args()
}
