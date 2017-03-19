package main

import "../gfx"
import "github.com/veandco/go-sdl2/sdl"
import "fmt"
import "sort"
import "math/rand"

func main() {
  tree := &gfx.RTree{Rect:sdl.Rect{10,11,1000,900}}
  tree.Merge(&gfx.RTree{Rect:sdl.Rect{100,200,300,400}})
  fmt.Println(tree)
  
  tree2 := &gfx.RTree{Rect:sdl.Rect{110,211,12,13}, Payload:1}
  tree2.Merge(&gfx.RTree{Rect:sdl.Rect{100,200,300,400}, Payload:2})
  tree2.Merge(&gfx.RTree{Rect:sdl.Rect{90,200,10,11}, Payload:3})
  fmt.Println(tree2)

  tree.Merge(tree2)
  tree.Merge(&gfx.RTree{Rect:sdl.Rect{100,200,300,400}})
  fmt.Println(tree)
  
  tree3 := &gfx.RTree{Rect:sdl.Rect{80,210,90,99}}
  tree3.Merge(&gfx.RTree{Rect:sdl.Rect{85,211,7,13},Payload:4})
  tree3.Merge(&gfx.RTree{Rect:sdl.Rect{85,222,8,14},Payload:5})
  fmt.Println(tree3)
  
  tree.Merge(tree3)
  
  fmt.Println(tree)
  
  tree = &gfx.RTree{Rect:sdl.Rect{0,0,100,100}}
  objs := make([]*gfx.RTree,100)
  typ := make([]int32,100)
  var x,y int32 
  for x = 0; x < 10; x++ {
    for y = 0; y < 10; y++ {
      var d int32 = (x+y)&1
      objs[y*10+x] = &gfx.RTree{Rect:sdl.Rect{x*10-d,y*10-d,10+d+d,10+d+d}, Payload:y*10+x+1}
      typ[y*10+x] = d
    }
  }
  
  for i := range objs {
    if typ[i] == 1 {
      tree.Merge(&gfx.RTree{Rect:objs[i].Rect, Payload:objs[i].Payload})
    }
  }
  
  for i := range objs {
    if typ[i] == 0 {
      tree.Merge(&gfx.RTree{Rect:objs[i].Rect, Payload:objs[i].Payload})
    }
  }

  /*
    Now we build tree2 to be a 5x5 grid of tree copies
  */
  ofs := []int{-2,-1,0,1,2}
  tree2 = &gfx.RTree{Rect:sdl.Rect{0,0,100,100}}
  for _, x := range ofs {
    for _, y := range ofs {
      dx := int32(x) * tree.W // tree, NOT tree2!!!
      dy := int32(y) * tree.H // tree, NOT tree2!!!
      for i := range objs {
        obj := &gfx.RTree{Rect:objs[i].Rect, Payload:objs[i].Payload}
        obj.X += dx
        obj.Y += dy
        tree2.Merge(obj)
      }
    }
  }
  
  /*
  Now we scroll the window across the 5x5 grid and query tree2 and
  compare the result to the same query against tree with wrap==true.
  */
  window := &sdl.Rect{tree2.X,tree2.Y,tree.W,tree.H}
  dx := tree.W >> 1
  dy := tree.H >> 1
  for window.Y+window.H <= tree2.Y+tree2.H {
    cmp := map[string]bool{}
    res := tree2.Query(window,false)
    for _,r := range res {
      cmp[fmt.Sprintf("%v %v",r.Rect,r.Payload)] = true
    }
    
    strs := make([]string,len(res))
    for i := range res { strs[i] = fmt.Sprintf("%v %v",res[i].Rect,res[i].Payload) }
    sort.Strings(strs)
    //fmt.Printf("nowrap %v %v\n",window,strs)
    
    res = tree.Query(window,true)
    
    strs = make([]string,len(res))
    for i := range res { strs[i] = fmt.Sprintf("%v %v",res[i].Rect,res[i].Payload) }
    sort.Strings(strs)
    //fmt.Printf("  wrap %v %v\n",window,strs)
    
    for _,r := range res {
      st := fmt.Sprintf("%v %v",r.Rect,r.Payload) 
      if !cmp[st] { panic("Wrapped query returned result "+st+" that shouldn't be there") }
      cmp[st] = false
    }
    
    for k := range cmp {
      if cmp[k] { panic("Wrapped query did not return "+k) }
    }
    
    window.X += dx
    if window.X < tree2.X || window.X+window.W > tree2.X+tree2.W {
      window.X -= dx
      window.Y += dy
      dx = -dx
    } 
  }
  
  

  fmt.Printf("Efficiency: %v%%  Std deviation: %v%%  Min: %v%%\n", gfx.QueryEfficiencyAverage(), gfx.QueryEfficiencyStdDeviation(), gfx.QueryEfficiencyMinimum())
  gfx.QueryEfficiencyReset()
  
  Screen := &sdl.Rect{0,0,1920,1080}
  var Sx int32 = 20 // #screens in x direction
  var Sy int32 = 5  // #screens in y direction
  
  lowshift := []uint{3, 5, 7}
  colorshift := []uint{8,16,24}
  var rects *gfx.RTree
  
  totalcount := 0
  
  for i := range lowshift {
    count := 0
    var y int32 = 0
    for y < Sy*Screen.H {
      var x int32 = 0
      for x < Sx*Screen.W {
        w := rnd(Screen.W >> lowshift[i], Screen.W >> (lowshift[i]-2))
        h := rnd(Screen.H >> lowshift[i], Screen.H >> (lowshift[i]-2))
        color := rand.Int31n(256) << colorshift[i]
        rects = &gfx.RTree{Rect:sdl.Rect{x,y,w,h}, Sibling:rects, Payload:color}
        x += rnd(Screen.W >> lowshift[i], Screen.W >> (lowshift[i]-2))
        count++
      }
      y += rnd(Screen.H >> lowshift[i], Screen.H >> (lowshift[i]-2))
    }
    fmt.Printf("Layer %v count: %v\n", i+1, count)
    totalcount += count
  }
  
  var World sdl.Rect
  for r := rects; r != nil; r = r.Sibling {
    World = r.Union(&World)
  }
  
  rects2 := &gfx.RTree{Rect:rects.Rect, Payload:rects.Payload}
  tree = rects
  rects = rects.Sibling
  tree.Sibling = nil 
  
  window = &sdl.Rect{Screen.X,Screen.Y,Screen.W,Screen.H}
  dx = Screen.W >> 2
  dy = Screen.H >> 2
  for window.Y < World.Y+World.H {
    var next *gfx.RTree
    prev := rects
    for r := rects; r != nil; r = next {
      next = r.Sibling
      if r.HasIntersection(window) {
        if r == rects {
          rects = r.Sibling
          prev = rects
        } else {
          prev.Sibling = r.Sibling
        }
        rects2 = &gfx.RTree{Rect:r.Rect, Payload:r.Payload, Sibling:rects2}
        r.Sibling = nil
        tree.Merge(r)
      } else {
        prev = r
      }
    }
    window.X += dx
    if window.X < World.X || window.X+window.W >= World.W {
      window.X -= dx
      window.Y += dy
      dx = -dx
    } 
  }
  
  if rects != nil { panic("foo") }
  rects = rects2
  for r := rects; r != nil; r = r.Sibling { totalcount-- }
  if totalcount != 0 { panic("bar") }
  
  if !tree.Rect.Equals(&World) { panic("foobar") }
  
  fmt.Println(tree.Rect)
  
  fmt.Println("Running query tests...")
  
  window = &sdl.Rect{Screen.X,Screen.Y,Screen.W,Screen.H}
  dx = Screen.W >> 3
  dy = Screen.H >> 3
  for window.Y+window.H < World.Y+World.H {
    res := tree.Query(window, false)
    rectsmap := map[string]bool{}
    
    for r := rects; r != nil; r = r.Sibling {
      if r.HasIntersection(window) {
        rectsmap[r.String()] = true
      }
    }
    
    for i := range res {
      st := res[i].String()
      if !rectsmap[st] {
        panic(fmt.Sprintf("Window: %v Result: %v not supposed to be there", window, res[i]))
      }
      rectsmap[st] = false
    }
    
    if len(rectsmap) != len(res) {
      for k := range rectsmap {
        if rectsmap[k] {
          panic(fmt.Sprintf("Window: %v Result: %v not returned by query", window, k))
        }
      }
    }
    
    window.X += dx
    if window.X < World.X || window.X+window.W > World.X+World.W {
      window.X -= dx
      window.Y += dy
      dx = -dx
      fmt.Printf("%v %v\n",window,gfx.QueryEfficiencyAverage())
    } 
  }
  
  fmt.Printf("Efficiency: %v%%  Std deviation: %v%%  Min: %v%%\n", gfx.QueryEfficiencyAverage(), gfx.QueryEfficiencyStdDeviation(), gfx.QueryEfficiencyMinimum())
   
}

func rnd(low, hi int32) int32 {
  return rand.Int31n(hi-low)+low
}
