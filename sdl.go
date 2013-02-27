package main

import (
	"./subtitle"
	"errors"
	"fmt"
	"github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"
	"github.com/0xe2-0x9a-0x9b/Go-SDL/ttf"
	"log"
	"strings"
	"time"
)

const (
	DEFAULT_FONT_PATH = "/usr/share/fonts/truetype/nanum/NanumGothicBold.ttf"
	DEFAULT_FONT_SIZE = 70
)

var (
	BG_COLOR    = sdl.Color{0, 0, 0, 0}
	TEXT_COLOR  = sdl.Color{255, 255, 255, 0}
	waitFinishC = make(chan bool)
)

type sdlCtx struct {
	surface *sdl.Surface
	w, h    int

	currScript *subtitle.Script
	font       *ttf.Font
	lineHeight int

	fontSize int
	fontPath string
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
	c.surface.FillRect(nil, 0 /* BG_COLOR */)
	c.surface.Flip()
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
	c.lineHeight = c.font.LineSkip()

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

	c.w, c.h = w, h
	if c.currScript != nil {
		c.displayScript(c.currScript, false, true)
	}

	return nil
}

func (c *sdlCtx) handelEvent() error {
	// KeySym to time.Duration mapping
	kmVias := map[uint32]time.Duration{
		sdl.K_z:       -100 * time.Microsecond,
		sdl.K_x:       +100 * time.Microsecond,
		sdl.K_LEFT:    -1 * time.Second,
		sdl.K_RIGHT:   +1 * time.Second,
		sdl.K_DOWN:    -10 * time.Second,
		sdl.K_UP:      +10 * time.Second,
		sdl.K_LESS:    -100 * time.Microsecond,
		sdl.K_GREATER: +100 * time.Microsecond,
	}

	kmFontSize := map[uint32]int{
		sdl.K_PLUS:     +5,
		sdl.K_KP_PLUS:  +5,
		sdl.K_MINUS:    -5,
		sdl.K_KP_MINUS: -5,
	}

	select {
	case event := <-sdl.Events:
		/* log.Printf("%#v\n", event) */
		switch e := event.(type) {
		case sdl.QuitEvent:
			return errors.New("sdl: received QuitEvent")
		case sdl.KeyboardEvent:
			// Ignore release key
			if e.State == 0 {
				return nil
			}
			// log.Printf("Sim:%08x, Mod:%04x, Unicode:%02x, %t\n",
			// 	e.Keysym.Sym, e.Keysym.Mod, e.Keysym.Unicode,
			// 	e.Keysym.Unicode)
			if vias, ok := kmVias[e.Keysym.Sym]; ok {
				viasC <- vias
				break
			}
			if vias, ok := kmFontSize[e.Keysym.Sym]; ok {
				c.changeFontSize(vias)
				break
			}
		case sdl.ResizeEvent:
			if err := c.setSurface(int(e.W), int(e.H)); err != nil {
				log.Fatal(err)
			}
		}
	}
	return nil
}

func (c *sdlCtx) displayScript(script *subtitle.Script,
	andClear bool, forceUpdate bool) {
	if forceUpdate == false && c.currScript == script {
		return
	}
	c.currScript = script

	log.Printf("display %s", script.Text)

	if c.font == nil {
		log.Println("set default font")
		err := c.setFont(DEFAULT_FONT_PATH, DEFAULT_FONT_SIZE)
		if err != nil {
			log.Fatal("failed to set default font")
			return
		}
	}

	c.surface.FillRect(nil, 0 /* BG_COLOR */)
	offsetX := 10
	offsetY := 10

	for _, line := range strings.Split(script.Text, "\n") {
		w, _, err := c.font.SizeUTF8(line)
		if err != 0 {
			log.Fatal("Failed to get size of the font")
		}
		offsetX = (c.w - w) / 2

		glypse := ttf.RenderUTF8_Blended(c.font, line, TEXT_COLOR)
		lt := sdl.Rect{int16(offsetX), int16(offsetY), 0, 0}
		c.surface.Blit(&lt, glypse, nil)
		offsetY += c.lineHeight

	}
	c.surface.Flip()

	if andClear == false {
		return
	}

	go func() {
		<-time.After(script.Duration())
		if c.currScript != nil && c.currScript == script {
			c.surface.FillRect(nil, 0 /* BG_COLOR */)
			c.surface.Flip()
		}
	}()
}
