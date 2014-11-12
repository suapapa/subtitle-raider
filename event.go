package main

import (
	"log"
	"os"
	"runtime"
	"time"

	"github.com/banthar/Go-SDL/sdl"
)

var (
	// KeySym to time.Duration mapping
	kmVias = map[uint32]time.Duration{
		sdl.K_z:     -100 * time.Microsecond,
		sdl.K_x:     +100 * time.Microsecond,
		sdl.K_LEFT:  -1 * time.Second,
		sdl.K_RIGHT: +1 * time.Second,
		sdl.K_DOWN:  -10 * time.Second,
		sdl.K_UP:    +10 * time.Second,
	}

	kmFontSize = map[uint32]int{
		sdl.K_EQUALS:   +5, // +
		sdl.K_MINUS:    -5,
		sdl.K_KP_PLUS:  +5,
		sdl.K_KP_MINUS: -5,
	}

	kmNavScript = map[uint32]int{
		sdl.K_SPACE:  0,
		sdl.K_COMMA:  -1, // <
		sdl.K_PERIOD: +1, // >
	}
)

func eventLoop(c *Screen) {
EVENTLOOP:
	/* log.Printf("%#v\n", event) */
	switch e := sdl.PollEvent().(type) {
	case *sdl.QuitEvent:
		os.Exit(0)

	case *sdl.ResizeEvent:
		if opt.fullscreen {
			break
		}
		if err := c.setSurface(int(e.W), int(e.H)); err != nil {
			log.Fatal(err)
		}
		c.updateC <- 1

	case *sdl.KeyboardEvent:
		// Ignore key-up
		if e.State == 0 {
			break
		}

		keysym := e.Keysym.Sym
		if keysym == sdl.K_q {
			quitC <- true
			break
		}

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
			navC <- v
			break
		}
		log.Printf("Sim:%08x, Mod:%04x, Unicode:%02x, %t\n",
			e.Keysym.Sym, e.Keysym.Mod, e.Keysym.Unicode,
			e.Keysym.Unicode)
	} // end of switch

	runtime.Gosched()
	goto EVENTLOOP
}
