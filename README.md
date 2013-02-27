

# Requirement

## GO-SDL

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

[Read more][1] about this issue.

Install Go-SDL:

    $ go get

It will install `github.com/0xe2-0x9a-0x9b/Go-SDL/sdl` and 
`github.com/0xe2-0x9a-0x9b/Go-SDL/ttf`.

[1]:https://github.com/banthar/Go-SDL/issues/35#issuecomment-3597261
