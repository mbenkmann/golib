package main

import "strings"
import "fmt"
import "../gfx"

var tests = []string{
`
XX XXX
 XoXX 
      
XXXXXX

`,
`
  o   
`,
`
    X  
  o    
       
       
    X  
`,
`
XXXXXX
XXXXXX
   o  
XXXXXX
XXXXXX
`,
`
   X  
X XX X
      
 XoXX 
XX XXX
`,
`
-------
---X---
--X X--
-X   X-
-Xo  X-
-XX  X-
-X  X--
-X X---
--X----
-------
`,
`
XX   o        
XXX  XX  XXX X
X          X X
             X
XXXXXXXXXXXX X
XXXXXXXXXXXX X
XXXXXXXXXXXX X
`,
`
XX   o    
XXX  XXX X
X      X X
         X
XXXXXXXX X
XXXXXXXX X
XXXXXXXX X
`,
`
XX   o    
X   X  X X
XX    XX X
         X
XXXXXXXX X
XXXXXXXX X
XXXXXXXX X
`,
`
-XX   o   
-X   XXX X
XXX    X X
         X
XXXXXXXX X
XXXXXXXX X
XXXXXXXX X
`,
`
XXX   o   
 X   XXX X
   XXXXX X
         X
XXXXXXXX X
XXXXXXXX X
XXXXXXXX X
`,
`
X X   o   
     XXX X
XX XXXXX X
X        X
XXXXXXXX X
XXXXXXXX X
XXXXXXXX X
`,
}


func main() {
  res := gfx.FloodFill(-1,0,10,10,func(int)bool{panic("oob")})
  if len(res) != 10 { panic("not 10") }
  for _, r := range res {
    if len(r) != 0 { panic("wtf") }
  }
  
  res = gfx.FloodFill(0,-1,10,10,func(int)bool{panic("oob")})
  if len(res) != 10 { panic("not 10") }
  for _, r := range res {
    if len(r) != 0 { panic("wtf") }
  }
  
  res = gfx.FloodFill(0,11,10,10,func(int)bool{panic("oob")})
  if len(res) != 10 { panic("not 10") }
  for _, r := range res {
    if len(r) != 0 { panic("wtf") }
  }
  
  res = gfx.FloodFill(11,0,10,10,func(int)bool{panic("oob")})
  if len(res) != 10 { panic("not 10") }
  for _, r := range res {
    if len(r) != 0 { panic("wtf") }
  }
  
  for _, test := range tests {
    fmt.Println()
    test = strings.Trim(test,"\n")
    lines := strings.Split(test,"\n")
    width := len(lines[0])
    height := len(lines)
    in := func(i int)bool{
      return lines[i/width][i%width] != 'X'
    }
    y := 0
    for !strings.Contains(lines[y],"o") {
      y++
    }
    x := strings.Index(lines[y],"o")
    res := gfx.FloodFill(x,y,width,height,in)
    if len(res) != height {
      panic("Result has wrong number of entries")
    }
    
    out := make([][]byte, len(lines))
    for y := 0; y < len(lines); y++ {
      out[y] = []byte(lines[y])
      for _, r := range res[y] {
        x1 := r & 0xFFFF
        x2 := r >> 16
        if x1 > x2 {
          panic("Range with x1 > x2")
        }
        for x1 <= x2 {
          if out[y][x1] == 'X' || out[y][x1] == '-' {
            panic("Coloring outside the lines")
          }
          out[y][x1] = 'X'
          x1++
        }
      }
      for i := range out[y] {
        if out[y][i] != 'X' && out[y][i] != '-' {
          panic(string(out[y]))
        }
      }
      fmt.Println(string(out[y]))
    }
  }
}
