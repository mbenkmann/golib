/* Copyright (C) 2017 Matthias S. Benkmann
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this file (originally named deb.go) and associated documentation files 
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

// Various re-usable utility functions.
package util

import (
         "strings"
         "strconv"
       )

// Returns true if v1 is older than v2 according to Debian version
// comparison rules.
//   See https://www.debian.org/doc/debian-policy/ch-controlfields.html
//   section "5.6.12 Version"
func DebVersionLess(v1,v2 string) bool {
  return DebVersionCompare(v1,v2) < 0
}

// Returns -1, 0, +1 if v1 is less, equal, greater than v2 according to Debian version
// comparison rules.
//   See https://www.debian.org/doc/debian-policy/ch-controlfields.html
//   section "5.6.12 Version"
func DebVersionCompare(v1,v2 string) int {
  epoch1, v1, rev1 := debSplit(v1)
  epoch2, v2, rev2 := debSplit(v2)
  if epoch1 < epoch2 { return -1 }
  if epoch2 < epoch1 { return +1 }
  if v1 == v2 {
    return debVersionPartCompare(rev1,rev2)
  }
  return debVersionPartCompare(v1,v2)
}

func debSplit(v1 string) (epoch, vers, rev string) {
  epoch1i := strings.Index(v1, ":")
  epoch = "0"
  if epoch1i > 0 {
    epoch = v1[0:epoch1i]
    v1 = v1[epoch1i+1:]
  }
  rev1i :=  strings.LastIndex(v1, "-")
  rev = "0"
  if rev1i >= 0 {
    rev = v1[rev1i+1:]
    v1 = v1[0:rev1i]
    if rev == "" { rev = "0" }
  }
  vers = v1
  return epoch, vers, rev
}

func debVersionPartCompare(v1,v2 string) int {
  i := 0
  for {
    c1 := -256 // end of string sorts before 0
    c2 := -256
    if i < len(v1) {
      c1 = int(v1[i])
    }
    if i < len(v2) {
      c2 = int(v2[i])
    }
    if c1 == int('~') {
      c1 = -512 // '~' sorts before everything, even end of string
    }
    if c2 == int('~') {
      c2 = -512 // '~' sorts before everything, even end of string
    }
    if (c1 >= int('A') && c1 <= int('Z')) || (c1 >= int('a') && c1 <= int('z')) {
      c1 -= 128 // letters sort before non-letters
    }
    if (c2 >= int('A') && c2 <= int('Z')) || (c2 >= int('a') && c2 <= int('z')) {
      c2 -= 128 // letters sort before non-letters
    }
    
    if c1 >= int('0') && c1 <= int('9') {
      c1 = -192 // digit is end of part which sorts after end of string
    }
    if c2 >= int('0') && c2 <= int('9') {
      c2 = -192 // digit is end of part which sorts after end of string
    }
    
    if c1 == -192 && c2 == -192 { // number compare required
      k1 := i
      for k1 < len(v1) && v1[k1] >= '0' && v1[k1] <= '9' { k1++ }
      k2 := i
      for k2 < len(v2) && v2[k2] >= '0' && v2[k2] <= '9' { k2++ }
      num1, e1 := strconv.Atoi(v1[i:k1])
      num2, e2 := strconv.Atoi(v2[i:k2])
      if e1 == nil && e2 == nil {
        c1 = num1
        c2 = num2
        // if k1 != k2 usually c1 != c2, if not for leading 0s, so
        // we will exit the loop and i doesn't matter.
        // if there are leading 0s, the comparison result will be messed
        // up, but who cares. GIGO.
        i = k1-1 // -1 because we increment later
      }
    }
    
    if c1 < c2 { return -1 }
    if c2 < c1 { return +1 }
    
    if i >= len(v1) { break }
    
    i++ // increment after the test against len(v1) because we want to execute the loop for len(v1)!
  }
  return 0
}

