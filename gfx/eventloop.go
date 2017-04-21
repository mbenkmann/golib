/* Copyright (C) 2017 Matthias S. Benkmann
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this file (originally named eventloop.go) and associated documentation files 
 * (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is furnished
 * to do so, subject to the following conditions:
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 * 
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE. 
 */

// Graphics related algorithms and data structures.
package gfx

import "time"
import "github.com/veandco/go-sdl2/sdl"

const (
  // Time per frame.
  FrameTime time.Duration = (time.Second+30)/60
  
  // If this number of dropped frames is exceeded, we assume that the process
  // has been suspended for some reason and will not try to catch up.
  MaxDroppedFrames = 10
)

// Handles an sdl.Event. Used with EventLoop().
type EventHandler interface{
  // Handle the event ev. Return values:
  //   consume: true => do not pass even to other handlers in the chain
  //   quit:    true => exit the event loop (immediately if consume == true)
  Handle(event sdl.Event) (consume, quit bool)
}

// Computes and renders frames. Used with EventLoop().
type Framer interface {
  // Compute frame with number num. The EventLoop() function guarantees that
  // this function is called for each number starting with 1 with no gaps
  // in the count. Note that multiple calls to ComputeFrame() may occur with
  // no calls to RenderFrame() in between (or at all). IOW ComputeFrame() must
  // not depend on anything RenderFrame() does.
  ComputeFrame(num int64)
  // Render the most recently computed frame. Note that while calls to ComputeFrame()
  // are never dropped, calls to RenderFrame() may be. Heavy stuff should
  // be done in RenderFrame() if possible.
  // NOTE: renderer.Clear() and renderer.Present() are called by EventLoop(). Do
  // not call them here unless you have some special (weird) reason.
  RenderFrame(renderer *sdl.Renderer)
}


// An EventHandler that causes EventLoop() to quit immediately (i.e. without
// further event processing) on the following events:
//   - sdl.QuitEvent
var Quitter = &quitter{}

// Set by EventLoop().
// "lost" frames count gaps in the program's execution that are so large that
// no attempt was made to catch up by dropping frames. This happens typically
// when the program is not scheduled for a while, e.g. because of SIGSTOP. In
// this case EventLoop() counts these frames as lost and simply continues with
// the next frame as if nothing happened.
var LostFrames int64

// Set by EventLoop().
// Dropped frames are frames that the program computes without presenting
// to the user in order to catch up with the frame that's supposed to be
// showing according to execution time.
var DroppedFrames int64

// Runs an SDL event loop that processes events once per frame as determined
// by FrameTime and MaxDroppedFrames. Each event is passed to each entry in
// evhandlers in turn until one of them returns consume==true.
// After all events are handled for the frame, the ComputeFrame() function
// will be called for each of the framers. If necessary to catch up with
// the frame time the framers list will be looped through several times,
// each time increasing the number passed to ComputeFrame().
// Once all calls to ComputeFrame() are done, the framers list is walked
// through ONCE and the RenderFrame() function is called for each framer.
// Entries of evhandlers that also support the Framer interface will be
// implicitly added to the BEGINNING of the framers list. Do not manually
// add them to framers or they will be called twice.
// When EventLoop() returns (because one of evhandlers has returned quit==true),
// you can find statistics in DroppedFrames and LostFrames.
func EventLoop(renderer *sdl.Renderer, evhandlers []EventHandler, framers []Framer) {
  t0 := time.Now()

  LostFrames = 0
  DroppedFrames = 0
  
  var lastframe int64 = 0
  
  running := true
  for {
    t := time.Since(t0)
    frame := int64((t+(FrameTime>>1)) / FrameTime) - LostFrames
    if frame - lastframe > MaxDroppedFrames+1 {
      LostFrames += frame - lastframe - 1
      frame = lastframe + 1
    }
    
    if frame == lastframe {
      time.Sleep(time.Millisecond)
      continue
    }
    
    DroppedFrames += frame - lastframe - 1
    
    for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
      for _, handler := range evhandlers {
        consume,quit := handler.Handle(event)
        if quit { running = false }
        if consume { break }
      }
    }
    
    if !running { break }
    
    for lastframe < frame {
      lastframe++
      for i := range evhandlers {
        framer, ok := evhandlers[i].(Framer)
        if ok { framer.ComputeFrame(lastframe) }
      }
      for i := range framers {
        framers[i].ComputeFrame(lastframe)
      }
    }

    renderer.Clear()
    
    for i := range evhandlers {
      framer, ok := evhandlers[i].(Framer)
      if ok { framer.RenderFrame(renderer) }
    }
    for i := range framers {
      framers[i].RenderFrame(renderer)
    }
    
    renderer.Present()
  }
}

type quitter struct{}

func (*quitter) Handle(event sdl.Event) (consume, quit bool) {
  switch event.(type) {
    case *sdl.QuitEvent: return true, true
  }
  return false,false
}
