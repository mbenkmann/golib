package main

import "time"
import "runtime"
import "../gfx"
import "github.com/veandco/go-sdl2/sdl"

const (
  winTitle       = "Test"
  winWidth       = 960
  winHeight      = 540
  logWidth       = 1920
  logHeight      = 1080
)

const r = uint32((128 << 24) + (255 << 16))
const g = uint32((128 << 24) + (255 << 8))
const b = uint32((128 << 24) + (255 << 0))

var letter_r = []uint32{
r,r,r,r,0,
r,0,0,r,0,
r,0,0,r,0,
r,r,r,0,0,
r,0,0,r,0,
r,0,0,0,r,
}

var letter_g = []uint32{
0,g,g,g,0,
g,0,0,0,g,
g,0,0,0,0,
g,0,g,g,g,
g,0,0,0,g,
0,g,g,g,g,
}

var letter_b = []uint32{
b,b,b,b,0,
b,0,0,0,b,
b,b,b,b,0,
b,0,0,0,b,
b,0,0,0,b,
b,b,b,b,0,
}


var texture_r *sdl.Texture
var texture_g *sdl.Texture
var texture_b *sdl.Texture

func main() {
  var window *sdl.Window
  var renderer *sdl.Renderer
  var event sdl.Event
  var running bool
  var err error
  runtime.LockOSThread()
  if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
    panic(err)
  }
  defer sdl.Quit()
  window, err = sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight, sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)
  if err != nil {
    panic(err)
  }
  defer window.Destroy()
  
  renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED/*|sdl.RENDERER_PRESENTVSYNC*/)
  if err != nil {
    panic(err)
  }
  
  err = renderer.SetLogicalSize(logWidth, logHeight)
  if err != nil {
    panic(err)
  }
  
  texture_r = gfx.TextureFromRGBA(renderer, letter_r, 5, 6)
  texture_g = gfx.TextureFromRGBA(renderer, letter_g, 5, 6)
  texture_b = gfx.TextureFromRGBA(renderer, letter_b, 5, 6)

  running = true
  for running {
    time.Sleep(15*time.Millisecond)
    
    for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
      switch event.(type) {
      case *sdl.QuitEvent:
        running = false
      }
    }
    
    renderFrame(renderer)
  }
  
}

func renderFrame(renderer *sdl.Renderer) {
  renderer.SetDrawColor(255,255,255,255)
  renderer.Clear()
  renderer.Copy(texture_r, nil, &sdl.Rect{0,0,500,500})
  renderer.Copy(texture_g, nil, &sdl.Rect{150,150,500,500})
  renderer.Copy(texture_b, nil, &sdl.Rect{300,300,500,500})
  renderer.Present()
}
