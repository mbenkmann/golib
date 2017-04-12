/* Copyright (C) 2017 Matthias S. Benkmann
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this file (originally named json.go) and associated documentation files 
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

import "strings"

// Converts s to JSON. The format of s is illustrated by the following example:
//   key1: "string value\n2nd line. Embedded \"quotes\"."
//   "key2" = 42
//   key3: true
//     # comment
//   key4: Hi, this works without "..."
//   key5: { "foo":"bar", "bla": 99
//           key5: ["bla","fasel"] }
//
// Compared to regular JSON you may optionally use the following shorthands:
//   * The surrounding {...} around the whole structure may be omitted.
//   * If a line starts with a key, the "..." surrouding the key may be omitted.
//   * The character '=' may be used instead of ':'.
//   * Everything following '#' (if not inside "...") is ignored till the end of line.
//   * A ',' as last character of the line may be omitted.
//   * If the first character after ':'/'=' is not one of -+.0123456789{[" and
//     the first word does not start with "true", "false" or "null", then it is
//     assumed to be a string that extends to the end of the line.
func AlmostJSON(s string) []byte {
  out := make([]byte,0,len(s))
  
  suppress_braces := false
  
  i := 0
  for i < len(s) && s[i] <= ' ' { i++ } // skip whitespace in front
  // s[i] is first non-white character
  if i < len(s) && s[i] == '{' {
    suppress_braces = true
  } else {
    out = append(out, '{')
  }

  comma := -1 // -1 => start, 0 => idle, 1 => seen ':' (i.e. comma possible), 2 => seen '\n' after ':' (i.e. comma necessary if word follows)
  
  for ; i < len(s); i++ {
    c := s[i]
    switch c {
      case '#': for i < len(s)-1 {
                  if s[i+1] == '\n' { break }
                  i++
                }
                continue
      case ',': comma = 0 // no double commas
      case ':',']','}': comma = 1
      case '=': comma = 1
                c = ':'
      case '\n' : if comma == 1 { comma = 2 }
                
    }
    
    if c <= ' ' { continue }
    
    
    if comma == 2 || comma < 0 {
      if comma > 0 && c != ']' && c != '}' {
        out = append(out, ',')
        comma = 0
      }
      if c == '_' || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
        comma = 0
        out = append(out, '"')
        for i < len(s)-1 && c != ':' && c != '=' && c > ' ' {
          out = append(out,c)
          i++
          c = s[i]
        }
        out = append(out, '"')
        i--
        continue
      }
    }
    
    if c == '"' { // pass through strings unchanged
      for i < len(s)-2 { // (test for len(s)-2 ensures we can do i++ twice)
        out = append(out, c)
        i++
        c = s[i]
        if c == '\\' {
          out = append(out, c)
          i++
          c = s[i]
        } else if c == '"' {
          break
        }
      }
    } else if c == ':' && i < len(s)-1 {
      i++
      for i < len(s) && s[i] <= ' ' { i++ }
      if i == len(s) || strings.IndexByte("-+.0123456789{[\"", s[i]) >= 0 || 
         (i+4 < len(s) && (s[i:i+4] == "null" || s[i:i+4] == "true")) ||
         (i+5 < len(s) && s[i:i+5] == "false") {
        
        // Probably not a string; or already has quotes. Don't auto-insert "..."
        
      } else {
        out = append(out,c,'"')
        for i < len(s) {
          c = s[i]
          if c == '\n' { break }
          if c == '"' || c == '\\' {
            out = append(out,'\\') // escape " and \ in string
          }
          out = append(out,c)
          i++
        }
        c = '"'
      }
      
      i--
    }
    
    out = append(out,c)
  }
  
  if !suppress_braces {
    out = append(out,'}')
  }
  return out
}
