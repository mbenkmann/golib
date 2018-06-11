/* Copyright (C) 2017 Matthias S. Benkmann
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this file (originally named floodfill.go) and associated documentation files 
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

import "unsafe"
import "winterdrache.de/bindings/sdl"

const (
  // Mask the red bits in an RGBA uint32.
  RMask uint32 = 255 << 16
  // Mask the green bits in an RGBA uint32.
  GMask uint32 = 255 << 8
  // Mask the blue bits in an RGBA uint32.
  BMask uint32 = 255 << 0
  // Mask the alpha bits in an RGBA uint32.
  AMask uint32 = 255 << 24
)

// Takes a slice of width*height RGBA pixels and creates a Texture from it.
// Format of pixel is:
//   bits 24..31 A (255 is fully opaque)
//   bits 16..23 R
//   bits 8..15  G
//   bits 0..7   B
func TextureFromRGBA(renderer *sdl.Renderer,image []uint32, width, height int) *sdl.Texture {
  surface := sdl.CreateRGBSurfaceFrom((*(*[999999999]byte)(unsafe.Pointer(&image[0])))[0:width*height*4], width, height, 32, int(width)<<2, RMask, GMask, BMask, AMask)
  if surface == nil { panic(sdl.GetError()) }
  defer surface.Free()
  t := renderer.CreateTextureFromSurface(surface)
  if t == nil { panic(sdl.GetError()) }
  return t
}

