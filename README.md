# subtitle-raider: A subtitle only player

Play subtitle on -it's own separate- screen according to is's timestamp.
It's useful to movie player which don't have subtitle capability
(like DLNA player).

> Contribute to my mother who complained about
> absence of Korean subtitle on a movie, TombRaider.

## How to build

[Install Go][1]

### Install requirement packages for Go-SDL (on Ubuntu 12.04)

Install dependency packages:

    $ sudo apt-get install libsdl1.2-dev libsdl-mixer* libsdl-image* libsdl-ttf*

On under Ubuntu 12.10, need to make `SDL_ttf.pc` to `/usr/lib/pkgconfig` with
following context:

    prefix=@prefix@
    exec_prefix=@exec_prefix@
    libdir=@libdir@
    includedir=@includedir@

    Name: SDL_ttf
    Description: ttf library for Simple DirectMedia Layer with FreeType 2 support
    Version: @VERSION@
    Requires: sdl >= @SDL_VERSION@
    Libs: -L${libdir} -lSDL_ttf
    Cflags: -I${includedir}/SDL

[Read more][2] about this issue.

### Build and Install

    $ go get
    $ go build
    $ go install

## Usage

Run with your subtitle. Currently only SRT is supported

    $ subtitle-raider example.srt

### Keybinding

* `<` and `>`           : Set to prev/next script
* `+` and `-`           : increase/decrease font size by 5
* `left` and `right`    : Adjust vias to -/+ 1 seconds
* `up` and `down`       : Adjust vias to -/+ 10 seconds
* `z` and `x`           : Adjust vias to -/+ 0.1 seconds
* `q`                   : Quit


[1]:http://golang.org/doc/install
[2]:https://github.com/banthar/Go-SDL/issues/35#issuecomment-3597261
