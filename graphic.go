// Copyright 2013, Homin Lee. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/banthar/Go-SDL/sdl"
	"github.com/banthar/Go-SDL/ttf"
	"github.com/suapapa/go_subtitle"
)

var (
	BG_COLOR         = sdl.Color{0, 0, 0, 0}
	TEXT_COLOR       = sdl.Color{255, 255, 255, 0}
	DEBUG_TEXT_COLOR = sdl.Color{32, 32, 32, 0}
	waitFinishC      = make(chan bool)
)

type Screen struct {
	surface *sdl.Surface
	fps     int
	w, h    uint16

	currScript *subtitle.Script
	lineHeight uint16

	fontSize int
	fontPath string
	font     *ttf.Font

	debugFont       *ttf.Font
	debugLineHeight uint16

	updateC chan int
}

func NewScreen(w, h int) *Screen {
	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		log.Fatal("failed to init sdl", sdl.GetError())
		return nil
	}

	if ttf.Init() != 0 {
		log.Fatal("failed to init ttf", sdl.GetError())
		return nil
	}

	var ctx Screen
	if err := ctx.setSurface(w, h); err != nil {
		log.Fatal(err)
	}

	var vInfo = sdl.GetVideoInfo()
	log.Println("HW_available = ", vInfo.HW_available)
	log.Println("WM_available = ", vInfo.WM_available)
	log.Println("Video_mem = ", vInfo.Video_mem, "kb")

	/* title := "Subtitle Player" */
	title := os.Args[0]
	icon := "" // path/to/icon
	sdl.WM_SetCaption(title, icon)

	sdl.EnableUNICODE(1)

	ctx.fps = opt.fps

	err := ctx.setFont(opt.fontPath, opt.fontSize)
	if err != nil {
		log.Fatal("failed to set default font")
		return nil
	}

	ctx.debugFont = ttf.OpenFont(DFLT_FONT_PATH, 20)
	if ctx.debugFont == nil {
		errMsg := fmt.Sprintf("failed to open debug font: %s",
			sdl.GetError())
		/* return errors.New(errMsg) */
		log.Fatal(errMsg)
	}
	ctx.debugLineHeight = uint16(ctx.debugFont.LineSkip())

	ctx.updateC = make(chan int)

	return &ctx
}

func (c *Screen) Release() {
	if c.font != nil {
		c.font.Close()
	}
	if c.surface != nil {
		c.surface.Free()
	}
	ttf.Quit()
	sdl.Quit()

	/* log.Printf("sdl Released...") */
}

func (c *Screen) Clear() {
	log.Println("clear")
	c.surface.FillRect(&sdl.Rect{0, int16(c.debugLineHeight), c.w, c.h}, 0 /* BG_COLOR */)
	c.updateC <- 1
	c.currScript = nil
}

func (c *Screen) setFont(path string, size int) error {
	if c.font != nil {
		c.font.Close()
	}

	if size < 10 {
		size = 10
	}
	c.font = ttf.OpenFont(path, size)
	if c.font == nil {
		errMsg := fmt.Sprintf("failed to open font from %s: %s",
			path, sdl.GetError())
		return errors.New(errMsg)
	}
	/* c.font.SetStyle(ttf.STYLE_UNDERLINE) */

	c.fontSize = size
	c.fontPath = path
	c.lineHeight = uint16(c.font.LineSkip())

	log.Printf("fontsize=%d lineheight=%d\n", c.fontSize, c.lineHeight)
	return nil
}

func (c *Screen) changeFontSize(by int) {
	if c.font == nil {
		return
	}
	c.setFont(c.fontPath, c.fontSize+by)
	c.DisplayScript(c.currScript)
}

func (c *Screen) setSurface(w, h int) error {
	log.Printf("setSurface to %dx%d", w, h)
	if opt.fullscreen {
		c.surface = sdl.SetVideoMode(w, h, 32, sdl.FULLSCREEN)
	} else {
		c.surface = sdl.SetVideoMode(w, h, 32, sdl.RESIZABLE)
	}
	if c.surface == nil {
		errMsg := fmt.Sprintf("sdl: failed to set video to %dx%d: %s",
			w, h, sdl.GetError())
		return errors.New(errMsg)
	}

	c.w, c.h = uint16(w), uint16(h)
	c.DisplayScript(c.currScript)

	return nil
}

func graphicLoop(c *Screen) {
	fpsTicker := time.NewTicker(time.Second / time.Duration(c.fps)) // 30fps
	dirtyCnt := 0

GRAPHIC_LOOP:
	select {
	case u := <-c.updateC:
		dirtyCnt += u
	case <-fpsTicker.C:
		if dirtyCnt > 0 {
			if dirtyCnt > 1 {
				log.Println("Lost some frame?. dirtyCnt =", dirtyCnt)
			}
			c.surface.Flip()
			dirtyCnt = 0
		}
	}
	goto GRAPHIC_LOOP
}

func (c *Screen) displayDebug(text string) {
	c.surface.FillRect(&sdl.Rect{0, 0, c.w, c.debugLineHeight}, 0 /* BG_COLOR */)
	glypse := ttf.RenderUTF8_Solid(c.debugFont, text, DEBUG_TEXT_COLOR)
	c.surface.Blit(&sdl.Rect{0, 0, 0, 0}, glypse, nil)
	c.updateC <- 1
	/* c.surface.Flip() */
}

func (c *Screen) DisplayScript(script *subtitle.Script) {
	if script == nil {
		return
	}

	c.currScript = script

	log.Printf("display %d.%s", script.Idx, script.TextWithoutMarkup())

	c.surface.FillRect(&sdl.Rect{0, int16(c.debugLineHeight), c.w, c.h}, 0 /* BG_COLOR */)
	offsetY := c.debugLineHeight

	for _, line := range strings.Split(script.TextWithoutMarkup(), "\n") {
		if strings.TrimSpace(line) == "" {
			offsetY += c.lineHeight
			continue
		}
		runeLine := []rune(line)
		runeLineLen := len(runeLine)
		runeLineStart := 0

		for runeLineStart != runeLineLen {
			/* log.Println("start =", runeLineStart, "len =", runeLineLen) */
			runeSubLine := runeLine[runeLineStart:]
			runeSubLineLen := len(runeSubLine)
			i := sort.Search(runeSubLineLen, func(i int) bool {
				w, _, _ := c.font.SizeUTF8(string(runeSubLine[:i]))
				return w+20 >= int(c.w)
			})
			/* log.Printf("runeSubLine=%s, i=%d\n", string(runeSubLine), i) */

			if i > runeSubLineLen {
				i = runeSubLineLen
			}

			w, _, _ := c.font.SizeUTF8(string(runeSubLine[:i]))
			for w > int(c.w) {
				i -= 1
				w, _, _ = c.font.SizeUTF8(string(runeSubLine[:i]))
			}

			subline := string(runeLine[runeLineStart : runeLineStart+i])
			subline = strings.TrimSpace(subline)
			/* log.Println("subline=", subline) */
			runeLineStart += i
			if runeLineStart > runeLineLen {
				runeLineStart = runeLineLen
			}

			if subline == "" {
				continue
			}

			offsetX := 10
			if opt.alignCenter {
				w, _, err := c.font.SizeUTF8(subline)
				if err != 0 {
					log.Fatal("Failed to get size of the font")
				}
				offsetX = (int(c.w) - w) / 2
			}

			glypse := ttf.RenderUTF8_Blended(c.font, subline, TEXT_COLOR)
			if glypse == nil {
				continue
			}
			lt := sdl.Rect{int16(offsetX), int16(offsetY), 0, 0}
			c.surface.Blit(&lt, glypse, nil)
			offsetY += c.lineHeight
		}
	}
	c.updateC <- 1
}
