/* Copyright (C) 2017 Matthias S. Benkmann
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this file (originally named rtree.go) and associated documentation files 
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

import ( 
         "fmt"
         "math"
         "strings"
         "github.com/veandco/go-sdl2/sdl"
       )

// How many children an RTree node is allowed to have.
// This number may be exceeded if the Merge() algorithm decides it is
// beneficial.
const RTREE_NODE_SIZE = 16

/*
  A tree for storing SDL rectangles. Note that this is not one of the common
  types of R-Tree. It has the following properties:

    - It is NOT balanced. The distance from root to leaf may vary.
    - Payload may be stored in inner nodes as well as leaves.
    - Nodes that carry Payload may not be resized (because the Payload is assumed
      to apply to the exact rectangle). If such a node needs to
      grow a parent node will be introduced.
    - When growing nodes the tree favors minimizing circumference over minimizing area.
      This favors more square rectangles.
    - Node overflows are handled by merging children, not by splitting the node.
    - When a node overflows Merge() examines all pairs of children to find
      the "best" pair to merge. This means insertion is O(n²).
    - It's completely based on my intuition about what would be a good data
      structure for my particular use case (managing a computer game world made up of
      images of different sizes). There's no science behind it.
      The central idea is to have the tree consist mostly of the actual rectangles
      from the component images that form the world. E.g. if there is a house with a
      tree in front with a flower in front, it makes sense to represent this in
      the tree. That this makes the tree unbalanced is irrelevant since we are
      working in memory and following a child pointer is as fast as following a
      sibling pointer.
*/
type RTree struct {
  // If Payload != nil, this is the rectangle that applies to the Payload, e.g.
  // if the Payload is a texture these would be the coordinates within the world.
  // If Payload == nil, this rectangle is the bounding box of all child rectangles.
  sdl.Rect
  // And arbitrary Payload. The tree itself only cares if this is nil or not.
  // ATTENTION! If Payload != nil, this means Rect may not be modified!
  Payload interface{}
  // Link to sibling if there is one.
  Sibling *RTree
  // Link to first child if there is one.
  Child *RTree
}

// Counts and returns the number of children of this node. Slow function. Do not use in a loop.
func (r *RTree) ChildCount() int {
  count := 0
  for kid := r.Child; kid != nil; kid = kid.Sibling {
    count++
  }
  return count
}

// Returns a fancy visual representation of this tree.
func (r *RTree) String() string {
  var s []string
  r.str(&s,"")
  return strings.Join(s,"")
}

// Integrates tree into r. tree must be a root or things will go wrong.
func (r *RTree) Merge(tree *RTree) {
  if tree.Sibling != nil {
    panic("RTree.Merge() called with RTree that has a Sibling") // Does not work with current code
  }
  if tree.H == 0 || tree.W == 0 {
    panic("RTree.Merge() called with empty rectangle") // Untested case. Probably doesn't work because an empty tree would be seen as fully enclosed by r regardless of its X/Y coordinates.
  }
  
  sect, _ := r.Intersect(&tree.Rect)
  
  // If r fully enclosed by tree but not identical (i.e. r is truely smaller)
  if sect.W == r.W && sect.H == r.H && (r.H != tree.H || r.W != tree.W) { 
    // swap the contents and children of tree and r (but not Sibling!!!)
    r.Rect, tree.Rect, r.Payload, tree.Payload, r.Child, tree.Child = 
    tree.Rect, r.Rect, tree.Payload, r.Payload, tree.Child, r.Child
    // At this point if one rectangle is fully contained within the other, it is guaranteed
    // that r is the larger one. This prevents growing a rectangle unnecessarily (which might
    // even require inserting a new one).
  } 
  
  union := r.Union(&tree.Rect)
    
  // Grow r if necessary
  if union.W != r.W || union.H != r.H {
    // If r cannot be resized, insert a level
    if r.Payload != nil {
      new_r := &RTree{Rect:r.Rect, Payload:r.Payload, Sibling:nil, Child:r.Child}
      r.Payload = nil
      r.Child = new_r
    }
    // grow r
    r.Rect = union
  }

  // NOTE: tree.Sibling == nil !!!

  // If tree has a Payload, add tree as new child
  if tree.Payload != nil {
    tree.Sibling = r.Child
    r.Child = tree
  } else { // if tree has no Payload, add its children directly (i.e. merge tree and r)
    if r.Child == nil {
      r.Child = tree.Child
    } else {
      kid := r.Child
      for kid.Sibling != nil { kid = kid.Sibling }
      kid.Sibling = tree.Child
    }
  }
  
  // If we exceed the max. size, we need to merge some children
  r.mergeChildren()
}

// If wrap == true, the world is assumed to be infinite and a query
// rectangle that exceeds past one of the boundaries will wraparound to
// the other side (with coordinates of wrapped results fixed up to match
// the query rectangle).
// All of the returned RTrees are copies that have no siblings or children.
// Modifying them will not affect r.
func (r *RTree) Query(area *sdl.Rect, wrap bool) []*RTree {
  dx := int32(0)
  dy := int32(0)
  if wrap {
    dx = ((area.X - r.X) / r.W) * r.W
    dy = ((area.Y - r.Y) / r.H) * r.H
    area = &sdl.Rect{area.X-dx,area.Y-dy,area.W,area.H}
  }
  
  result, count := r.query(area, dx, dy)
  
  if wrap {
    rx1 := r.X
    ry1 := r.Y
    rx2 := rx1+r.W
    ry2 := ry1+r.H
    ax1 := area.X
    ay1 := area.Y
    ax2 := ax1+area.W
    ay2 := ay1+area.H
    // top-left
    shift := &sdl.Rect{0,0, area.W, area.H}
    if ax1 < rx1 && ay1 < ry1 {
      shift.X = rx2-(rx1-ax1)
      shift.Y = ry2-(ry1-ay1)
      res, c := r.query(shift, -shift.X+ax1+dx, -shift.Y+ay1+dy)
      result = append(result, res...)
      count += c
    }
    // top
    if ay1 < ry1 {
      shift.X = ax1
      shift.Y = ry2-(ry1-ay1)
      res, c := r.query(shift, dx, -shift.Y+ay1+dy)
      result = append(result, res...)
      count += c
    }
    // top-right
    if ax2 > rx2 && ay1 < ry1 {
      shift.X = ax1-r.W
      shift.Y = ry2-(ry1-ay1)
      res, c := r.query(shift, r.W+dx, -shift.Y+ay1+dy)
      result = append(result, res...)
      count += c
    }
    // left
    if ax1 < rx1 {
      shift.X = rx2-(rx1-ax1)
      shift.Y = ay1
      res, c := r.query(shift, -shift.X+ax1+dx, dy)
      result = append(result, res...)
      count += c
    }
    // right
    if ax2 > rx2 {
      shift.X = ax1-r.W
      shift.Y = ay1
      res, c := r.query(shift, r.W+dx, dy)
      result = append(result, res...)
      count += c
    }
    // bottom-left
    if ax1 < rx1 && ay2 > ry2 {
      shift.X = rx2-(rx1-ax1)
      shift.Y = ay1-r.H
      res, c := r.query(shift, -shift.X+ax1+dx, r.H+dy)
      result = append(result, res...)
      count += c
    }
    // bottom
    if ay2 > ry2 {
      shift.X = ax1
      shift.Y = ay1-r.H
      res, c := r.query(shift, dx, r.H+dy)
      result = append(result, res...)
      count += c
    }
    // bottom-right
    if ay2 > ry2 && ax2 > rx2 {
      shift.X = ax1-r.W
      shift.Y = ay1-r.H
      res, c := r.query(shift, r.W+dx, r.H+dy)
      result = append(result, res...)
      count += c
    }
  }
  
  // count can not be 0 because we called query() at least once
  // varianceShift see https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance
  // as protection against huge numbers (the precision argument does not apply here because
  // we use fixed point, not floating point)
  // should not actually be necessary because we should never exceed the range of int64,
  // but it doesn't hurt.
  efficiency := (int64(len(result))<<16)/int64(count) - varianceShift
  if efficiency < minEfficiency {
    minEfficiency = efficiency
  }
  efficiencySum += efficiency
  efficiency2Sum += efficiency*efficiency
  queryCount++
  
  return result
}

func (r *RTree) query(area *sdl.Rect, shiftX int32, shiftY int32) ([]*RTree,int) {
  if !r.HasIntersection(area) { return nil,1 }
  
  count := 1
  result := make([]*RTree,0,16)
  if r.Payload != nil {
    result = append(result, &RTree{Rect:sdl.Rect{r.X+shiftX,r.Y+shiftY,r.W,r.H},Payload:r.Payload})
  }
  
  for kid := r.Child; kid != nil; kid = kid.Sibling {
    res,c := kid.query(area, shiftX, shiftY)
    result = append(result, res...)
    count += c
  }
  
  return result, count
}

func (r *RTree) mergeChildren() {
  for r.ChildCount() > RTREE_NODE_SIZE {
    var best_kid1 *RTree
    var best_kid2 *RTree
    var best_cu int32 = 2147483647
    var best_mcc int
    
    for kid1 := r.Child; kid1 != nil; kid1 = kid1.Sibling {
      for kid2 := kid1.Sibling; kid2 != nil; kid2 = kid2.Sibling {
        union := kid1.Union(&kid2.Rect)
        // circumference of the union rectangle
        cu := union.H+union.W
        /*
        // The following code would optimize for circumference increase
        // but this favors creating big rectangles that swallow everything,
        // because once a rectangle is big it doesn't need to grow anymore
        // so its circumference increase is 0 most of the time.
        c1 := kid1.H+kid1.W
        c2 := kid2.H+kid2.W
        if c1 > c2 {
          cu -= c1
        } else {
          cu -= c2
        }*/
        
        // If circumference is worse, don't look further.
        if cu > best_cu { continue }
        
        mergeChildCount := 0
        if kid1.Payload != nil { 
          mergeChildCount++ 
        } else {
          mergeChildCount += kid1.ChildCount()
        }
        if kid2.Payload != nil { 
          mergeChildCount++ 
        } else {
          mergeChildCount += kid2.ChildCount()
        }
        
        if cu < best_cu || // we favor smaller circumference increase
           (cu == best_cu && mergeChildCount <= RTREE_NODE_SIZE && 
              (best_mcc > RTREE_NODE_SIZE || // we favor creating nodes that are not overfull
              // but as long as they are not overfull we prefer fuller nodes
               mergeChildCount > best_mcc)) {
          best_kid1 = kid1
          best_kid2 = kid2
          best_mcc = mergeChildCount
          best_cu = cu 
        }
      }
    }
    
    // Remove kid2 from the siblings list
    for k := best_kid1; true ; k = k.Sibling {
      if k.Sibling == best_kid2 {
        k.Sibling = best_kid2.Sibling
        best_kid2.Sibling = nil
        break
      }
    }
    
    // merge
    best_kid1.Merge(best_kid2)
  }
}

// use 60% as shift, the closer to the average, the better
const varianceShift = (60<<16)/100
var efficiencySum int64 = 0
var efficiency2Sum int64 = 0
var minEfficiency int64 = 999999999999
var queryCount int64 = 0

// Resets the efficiency counters used by QueryEfficiency...().
// Query efficiency is defined as
//   (number of nodes returned)/(number of nodes visited)
// An efficiency of 100% means that all nodes visited during query
// processing were part of the result set. Obviously efficiency is
// influenced both by the tree structure and the queries performed.
func QueryEfficiencyReset() {
  efficiencySum = 0
  efficiency2Sum = 0
  minEfficiency = 999999999999
  queryCount = 0
}

// Returns a percentage from 0 to 100.
func QueryEfficiencyAverage() int {
  if queryCount == 0 { return 0 }
  return int(((varianceShift + (efficiencySum/queryCount))*100) >> 16)
}

// Returns a percentage from 0 to 100.
func QueryEfficiencyMinimum() int {
  if queryCount == 0 { return 0 }
  return int(((varianceShift + minEfficiency)*100) >> 16)
}

func QueryEfficiencyVariance() int {
  if queryCount < 2 { return 0 }
  return int((((efficiency2Sum - (efficiencySum*efficiencySum)/queryCount)/(queryCount-1))*10000) >> 32)
}

func QueryEfficiencyStdDeviation() int {
  v := QueryEfficiencyVariance()
  return int(math.Sqrt(float64(v))+.5)
}

func (r *RTree) str(s *[]string, indent string) {
  start := len(*s)
  *s = append(*s, indent)
  
  // ┌─23,44(10,10)─
  // └─────33,54─┘
  
  upper := "╔"
  verti := "║"
  horiz := "═"
  lower := "╚"
  lowri := "╝"
  topri := "╗"
  if r.Payload != nil {
    upper = "┌"
    verti = "│"
    horiz = "─"
    lower = "└"
    lowri = "┘"
    topri = "┐"
  }
  
  header := fmt.Sprintf("%v,%v(%v,%v)",r.Rect.X,r.Rect.Y,r.Rect.W,r.Rect.H)
  payload := "foobar" // defines default payload length for non-payload rectangles
  *s = append(*s, upper, horiz, header, horiz)
  if r.Payload != nil {
    payload = fmt.Sprintf("%v%v%v",horiz,r.Payload,horiz)
    *s = append(*s, payload)
  }
  *s = append(*s, "", "\n")
  
  for kid := r.Child; kid != nil; kid = kid.Sibling {
    kid.str(s, indent+verti+" ")
  }
  
  footer := fmt.Sprintf("%v,%v",r.Rect.X+r.Rect.W,r.Rect.Y+r.Rect.H) 
  *s = append(*s, indent)
  *s = append(*s, lower)
  for i := 4+len([]rune(payload))+len([]rune(header)) - len([]rune(footer)); i > 0; i-- {
    *s = append(*s, horiz)
  }
  *s = append(*s, footer)
  *s = append(*s, "", "\n")
  
  ln := 0
  l := 0
  for i := start; i < len(*s); i++ {
    st := (*s)[i]
    if st == "\n" {
      if l > ln { ln = l }
      l = 0
    } else {
      l += len([]rune((*s)[i]))
    }
  }
  
  l = 0
  first := true
  for i := start; i < len(*s); i++ {
    if (*s)[i] == "\n" {
      last := (i+1 == len(*s))
      
      for l < ln {
        if first||last {
          (*s)[i-1] = (*s)[i-1] + horiz
        } else {
          (*s)[i-1] = (*s)[i-1] + " "
        }
        l++
      }
      
      if first {
        (*s)[i-1] =  (*s)[i-1] + topri
      } else if last {
        (*s)[i-1] = (*s)[i-1] + lowri
      } else {
        (*s)[i-1] = (*s)[i-1] + verti
      }

      l = 0
      first = false
    } else {
      l += len([]rune((*s)[i]))
    }
  }
}
