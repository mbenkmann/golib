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

/*
Scans an image of width*height pixels starting at x,y
with flood fill, where in() determines if a pixel is in the area to be filled.
in() is passed an index into a width*height array that is assumed to hold the
pixel data and is expected to return true if the respective pixel is to be filled.
Returns a slice with one entry for each y coordinate (i.e. height entries in total).
Each entry is a slice of uints that encode a part of the line to be filled as follows
  i & 0xFFFF  is the x coordinate of the first pixel to be colored
  i >> 16     is the x coordinate of the last pixel to be colored 
*/
func FloodFill(x,y, width, height int, in func(int)bool) [][]int {
  res := make([][]int, height)
  if x < 0 || x >= width || y < 0 || y >= height { return res }
  idx := width * y + x
  if !in(idx) { return res }
  x0 := x
  for x0 > 0 && in(idx-1) {
    x0--
    idx--
  }
  for x < width-1 && in(idx+1) {
    x++
    idx++
  }
  
  var todo [2][]int
  var dir = [2]int{-1,1}
  
  line := x0|(x<<16)
  res[y] = append(res[y], line)
  
  for i := 0; i <= 1; i++ {
    if y+dir[i] >= 0 && y+dir[i] < height {
      todo[i] = append(todo[i], y+dir[i], line)
    }
  }
  
  for len(todo[0])+len(todo[1]) > 0 {
    for i := 0; i <=1; i++ {
      if len(todo[i]) > 0 {
        line = todo[i][len(todo[i])-1]
        y = todo[i][len(todo[i])-2]
        todo[i] = todo[i][0:len(todo[i])-2]
        x0 = line & 0xFFFF
        x  = line >> 16
        x0, x = rangesubtract(x0, x, res[y], y, &todo[i])
        if x0 <= x {
          idx = y * width + x0
          a0 := x0
          aidx := idx
          if in(aidx) {
            for a0 > 0 && in(aidx-1) {
              a0--
              aidx--
            }
          }
          
          // if we extended past the initial line, we need to check other direction, too
          if a0 < x0 {
            if y+dir[1-i] >= 0 && y+dir[1-i] < height {
              todo[1-i] = append(todo[1-i], y+dir[1-i], a0|((x0-1) << 16))
            }
          } else {
            a0 = -1
          }
          
          for x0 <= x {
            if in(idx) {
              if a0 >= 0 {
                line = a0
                a0 = -1
              } else {
                line = x0
              }
              for x0 < width-1 && in(idx+1) {
                x0++
                idx++
              }
              line |= x0 << 16
              res[y] = append(res[y], line)
              if y+dir[i] >= 0 && y+dir[i] < height {
                todo[i] = append(todo[i], y+dir[i], line)
              }
              
              // if we extended past the initial line, we need to check other direction, too
              if x0 > x {
                if y+dir[1-i] >= 0 && y+dir[1-i] < height {
                  todo[1-i] = append(todo[1-i], y+dir[1-i], (x+1)|(x0 << 16))
                }
              }
              
            }
            x0++
            idx++
          }
        }
      }
    }
  }
  
  return res
}

/*
  Takes a range [x1,x2] and a list of ranges stored in lines as (start | end << 16)
  and subtracts all these ranges from [x1,x2]. There are 3 possibilities:
    1) the result is an empty range => returns a,b with a>b
    2) the result is a single range => returns a,b with a<=b
    3) the result are multiple disjoint ranges => returns one of them and appends
       to *overflow for each range y followed by (a|b<<16).
*/
func rangesubtract(x1,x2 int, lines []int, y int, overflow *[]int) (int,int) {
  ranges := []int{x1,x2}
  for _, l := range lines {
    l1 := int(l & 0xFFFF)
    l2 := int(l >> 16)
    
    for i := 0; i < len(ranges); i+=2 {
      x1 = ranges[i]
      x2 = ranges[i+1]
      
      if l2 < x1 || l1 > x2 { continue } // no intersection
      if l1 <= x1 {
        x1 = l2+1
      }
      if l2 >= x2 {
        x2 = l1-1
      }
      
      if x1 > x2 {
        ranges[i] = ranges[len(ranges)-2]
        ranges[i+1] = ranges[len(ranges)-1]
        ranges = ranges[0:len(ranges)-2]
        if len(ranges) == 0 {
          break
        } else {
          i -= 2
          continue
        }
      }
      
      if l1 > x1 && l2 < x2 {
        ranges = append(ranges, l2+1, x2)
        x2 = l1-1
      }
      
      ranges[i] = x1
      ranges[i+1] = x2
    }
  }
  
  if len(ranges) == 0 { return 1,0 }
  
  for i := 2; i < len(ranges); i+=2 {
    *overflow = append(*overflow, y, ranges[i]|(ranges[i+1]<<16))
  }
  
  return ranges[0],ranges[1]
}

