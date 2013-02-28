// Copyright 2013, Homin Lee. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"./subtitle"
	"errors"
	"fmt"
	"github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"
	"github.com/0xe2-0x9a-0x9b/Go-SDL/ttf"
	"log"
	"sort"
	"strings"
	"time"
)

const (
	DEFAULT_FONT_PATH = "/usr/share/fonts/truetype/nanum/NanumGothicBold.ttf"
	DEFAULT_FONT_SIZE = 90
)

var (
	BG_COLOR         = sdl.Color{0, 0, 0, 0}
	TEXT_COLOR       = sdl.Color{255, 255, 255, 0}
	DEBUG_TEXT_COLOR = sdl.Color{32, 32, 32, 0}
	waitFinishC      = make(chan bool)
)

type sdlCtx struct {
	surface      *sdl.Surface
	dirtySurface bool
	fps          int
	w, h         uint16

	currScript *subtitle.Script
	lineHeight uint16

	fontSize int
	fontPath string
	font     *ttf.Font

	debugFont       *ttf.Font
	debugLineHeight uint16
}

func NewSdlContext(w, h int) *sdlCtx {
	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		log.Fatal("failed to init sdl", sdl.GetError())
		return nil
	}

	if ttf.Init() != 0 {
		log.Fatal("failed to init ttf", sdl.GetError())
		return nil
	}

	var ctx sdlCtx
	if err := ctx.setSurface(w, h); err != nil {
		log.Fatal(err)
	}

	var vInfo = sdl.GetVideoInfo()
	log.Println("HW_available = ", vInfo.HW_available)
	log.Println("WM_available = ", vInfo.WM_available)
	log.Println("Video_mem = ", vInfo.Video_mem, "kb")

	title := "Subtitle Player"
	icon := "" // path/to/icon
	sdl.WM_SetCaption(title, icon)

	sdl.EnableUNICODE(1)

	// XXX: fix it to set from argument
	ctx.fps = 30

	ctx.debugFont = ttf.OpenFont(DEFAULT_FONT_PATH, 20)
	if ctx.debugFont == nil {
		errMsg := fmt.Sprintf("failed to open font from %s: %s",
			DEFAULT_FONT_PATH, sdl.GetError())
		/* return errors.New(errMsg) */
		log.Fatal(errMsg)
	}
	ctx.debugLineHeight = uint16(ctx.debugFont.LineSkip())

	go func() {
	EVENT_LOOP:
		for {
			err := ctx.handelEvent()
			if err != nil {
				fmt.Println(err)
				break EVENT_LOOP
			}
		}
		log.Println("sdl: exit event loop")
		quitC <- true
	}()

	return &ctx
}

func (c *sdlCtx) Release() {
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

func (c *sdlCtx) DisplayScript(script *subtitle.Script) {
	c.displayScript(script, true, false)
}

func (c *sdlCtx) Clear() {
	if c.currScript == nil {
		return
	}
	log.Println("clear")
	c.surface.FillRect(&sdl.Rect{0, int16(c.debugLineHeight), c.w, c.h}, 0 /* BG_COLOR */)
	c.dirtySurface = true
	/* c.surface.Flip() */
	c.currScript = nil
}

func (c *sdlCtx) setFont(path string, size int) error {
	if size < 10 {
		log.Println("set font size to minimum, 10")
		size = 10
	}

	c.font = ttf.OpenFont(path, size)
	if c.font == nil {
		errMsg := fmt.Sprintf("failed to open font from %s: %s",
			path, sdl.GetError())
		return errors.New(errMsg)
	}

	c.fontSize = size
	c.fontPath = path
	c.lineHeight = uint16(c.font.LineSkip())

	log.Println(c.fontSize, c.lineHeight)
	/* ctx.font.SetStyle(ttf.STYLE_UNDERLINE) */
	return nil
}

func (c *sdlCtx) changeFontSize(by int) {
	if c.font == nil {
		return
	}
	c.setFont(c.fontPath, c.fontSize+by)
	c.displayScript(c.currScript, false, true)
}

func (c *sdlCtx) setSurface(w, h int) error {
	log.Printf("setSurface to %dx%d", w, h)
	c.surface = sdl.SetVideoMode(w, h, 32, sdl.RESIZABLE) /* sdl.FULLSCREEN */
	if c.surface == nil {
		errMsg := fmt.Sprintf("sdl: failed to set video to %dx%d: %s",
			w, h, sdl.GetError())
		return errors.New(errMsg)
	}

	c.w, c.h = uint16(w), uint16(h)
	if c.currScript != nil {
		c.displayScript(c.currScript, false, true)
	}

	return nil
}

func (c *sdlCtx) handelEvent() error {
	// KeySym to time.Duration mapping
	kmVias := map[uint32]time.Duration{
		sdl.K_z:     -100 * time.Microsecond,
		sdl.K_x:     +100 * time.Microsecond,
		sdl.K_LEFT:  -1 * time.Second,
		sdl.K_RIGHT: +1 * time.Second,
		sdl.K_DOWN:  -10 * time.Second,
		sdl.K_UP:    +10 * time.Second,
	}

	kmFontSize := map[uint32]int{
		sdl.K_EQUALS:   +5, // +
		sdl.K_MINUS:    -5,
		sdl.K_KP_PLUS:  +5,
		sdl.K_KP_MINUS: -5,
	}

	kmNavScript := map[uint32]int{
		sdl.K_SPACE:  0,
		sdl.K_COMMA:  -1, // <
		sdl.K_PERIOD: +1, // >
	}

	fpsTicker := time.NewTicker(time.Second / time.Duration(c.fps)) // 30fps

	select {
	case <-fpsTicker.C:
		if c.dirtySurface {
			c.surface.Flip()
			c.dirtySurface = false
		}

	case event := <-sdl.Events:
		/* log.Printf("%#v\n", event) */
		switch e := event.(type) {
		case sdl.QuitEvent:
			return errors.New("sdl: received QuitEvent")

		case sdl.ResizeEvent:
			if err := c.setSurface(int(e.W), int(e.H)); err != nil {
				log.Fatal(err)
			}

		case sdl.KeyboardEvent:
			// Ignore release key
			if e.State == 0 {
				return nil
			}

			keysym := e.Keysym.Sym
			// tune timestamp
			if v, ok := kmVias[keysym]; ok {
				tsViasC <- v
				break
			}
			// tune font size
			if v, ok := kmFontSize[keysym]; ok {
				c.changeFontSize(v)
				break
			}

			// pause/resume
			if v, ok := kmNavScript[keysym]; ok {
				log.Println("TODO: nav tsVias=", v)
				navC <- v
				break
			}
			log.Printf("Sim:%08x, Mod:%04x, Unicode:%02x, %t\n",
				e.Keysym.Sym, e.Keysym.Mod, e.Keysym.Unicode,
				e.Keysym.Unicode)
		}
	}
	return nil
}

func (c *sdlCtx) displayDebug(text string) {
	c.surface.FillRect(&sdl.Rect{0, 0, c.w, c.debugLineHeight}, 0 /* BG_COLOR */)
	glypse := ttf.RenderUTF8_Solid(c.debugFont, text, DEBUG_TEXT_COLOR)
	c.surface.Blit(&sdl.Rect{0, 0, 0, 0}, glypse, nil)
	c.surface.Flip()
}

func (c *sdlCtx) displayScript(script *subtitle.Script,
	andClear bool, forceUpdate bool) {
	if forceUpdate == false && c.currScript == script {
		return
	}
	c.currScript = script

	log.Printf("display %d.%s", script.Idx, script.Text)

	if c.font == nil {
		log.Println("set default font")
		err := c.setFont(DEFAULT_FONT_PATH, DEFAULT_FONT_SIZE)
		if err != nil {
			log.Fatal("failed to set default font")
			return
		}
	}

	c.surface.FillRect(&sdl.Rect{0, int16(c.debugLineHeight), c.w, c.h}, 0 /* BG_COLOR */)
	offsetY := c.debugLineHeight

	for _, line := range strings.Split(script.TextWithoutMarkup(), "\n") {
		runeLine := []rune(line)
		runeLineLen := len(runeLine)
		runeLineStart := 0

		for runeLineStart != runeLineLen {
			/* log.Println("start =", runeLineStart, "len =", runeLineLen) */
			runeSubLine := runeLine[runeLineStart:]
			i := sort.Search(len(runeSubLine), func(i int) bool {
				w, _, _ := c.font.SizeUTF8(string(runeSubLine[:i]))
				return uint16(w-20) >= c.w
			})
			/* log.Println("runeSubLine=", string(runeSubLine)) */

			if i != len(runeSubLine) && i > 1 {
				i -= 1
			}
			if i > runeLineLen {
				i = runeLineLen
			}
			/* log.Println("returned i=", i) */

			subline := string(runeLine[runeLineStart : runeLineStart+i])
			subline = strings.TrimSpace(subline)
			/* log.Println("subline=", subline) */
			runeLineStart += i
			if runeLineStart > runeLineLen {
				runeLineStart = runeLineLen
			}

			w, _, err := c.font.SizeUTF8(subline)
			if err != 0 {
				log.Fatal("Failed to get size of the font")
			}
			offsetX := (c.w - uint16(w)) / 2

			glypse := ttf.RenderUTF8_Blended(c.font, subline, TEXT_COLOR)
			lt := sdl.Rect{int16(offsetX), int16(offsetY), 0, 0}
			c.surface.Blit(&lt, glypse, nil)
			offsetY += c.lineHeight
		}

	}
	c.dirtySurface = true
	/* c.surface.Flip() */

	if andClear == false {
		return
	}

	go func() {
		<-time.After(script.Duration())
		if c.currScript != nil && c.currScript == script {
			c.surface.FillRect(&sdl.Rect{0, int16(c.debugLineHeight), c.w, c.h}, 0 /* BG_COLOR */)
			c.surface.Flip()
		}
	}()
}
