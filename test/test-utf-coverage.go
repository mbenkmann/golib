package main

import (
         "io"
         "fmt"
         "errors"
         "../runes"
       )

var LATIN1 rune = 27362
var ASCII rune = 27363

type Test struct {
  Encoding int
  In []byte
  Out []rune
  Eof error 
}

type TestReader struct {
  test *Test
  index int
}

func (t *TestReader) Read(p []byte) (n int,err error) {
  if t.index > len(t.test.In)  {
    if t.test.Eof != nil {
      panic("Read past IO error")
    }
    return 0, io.EOF
  }
  
  for {
    if t.index == len(t.test.In) {
      t.index++
      if t.test.Eof == nil { return n,io.EOF }
      return n,t.test.Eof
    }
    
    if len(p) == 0 { return n,nil }
    
    p[0] = t.test.In[t.index]
    t.index++
    n++
    p = p[1:]
  }
}

var tests = []Test{
Test{ // 1
runes.ENCODING_UTF8,
[]byte{},
[]rune{runes.EOF, runes.EOF},
nil,
},

Test{ // 2
runes.ENCODING_8BIT,
[]byte{},
[]rune{LATIN1,runes.EOF},
nil,
},

Test{ // 3
runes.ENCODING_8BIT,
[]byte{},
[]rune{LATIN1,runes.UNSPECIFIC_IO_ERROR,runes.EOF},
errors.New("Foo"),
},

Test{ // 4
runes.ENCODING_UTF8,
[]byte{42},
[]rune{42, runes.EOF},
nil,
},

Test{ // 5
runes.ENCODING_UTF8,
[]byte{0xC0},
[]rune{runes.GARBAGE+0xC0, runes.EOF},
nil,
},

Test{ // 6
runes.ENCODING_UTF8,
[]byte{42},
[]rune{42, runes.UNSPECIFIC_IO_ERROR, runes.EOF},
errors.New("foo"),
},

Test{ // 7
runes.ENCODING_UTF8,
[]byte{0xC0},
[]rune{runes.GARBAGE+0xC0, runes.UNSPECIFIC_IO_ERROR, runes.EOF},
errors.New("foo"),
},

Test{ // 8
runes.ENCODING_UTF8,
[]byte{0xC2},
[]rune{runes.GARBAGE+0xC2, runes.EOF},
nil,
},

Test{ // 9
runes.ENCODING_UTF8,
[]byte{"Ä"[0], "Ä"[1]},
[]rune{'Ä', runes.EOF},
nil,
},

Test{ // 10
runes.ENCODING_UTF8,
[]byte{},
[]rune{runes.UNSPECIFIC_IO_ERROR,runes.EOF},
errors.New("Foo"),
},

Test{ // 11
runes.ENCODING_UTF8,
[]byte{'a'},
[]rune{'a',runes.UNSPECIFIC_IO_ERROR,runes.EOF},
errors.New("Foo"),
},

Test{ // 12
runes.ENCODING_UTF8,
[]byte{"Ä"[0]},
[]rune{runes.GARBAGE + rune("Ä"[0]),runes.UNSPECIFIC_IO_ERROR,runes.EOF},
errors.New("Foo"),
},

Test{ // 13
runes.ENCODING_UTF16LE,
[]byte{'a',0},
[]rune{'a',runes.EOF},
nil,
},

Test{ // 14
runes.ENCODING_UTF16BE,
[]byte{0,'a',1},
[]rune{'a',runes.GARBAGE+1,runes.EOF},
nil,
},

Test{ // 15
runes.ENCODING_UTF16LE,
[]byte{0xFF,0xFE},
[]rune{0xFEFF,runes.EOF},
nil,
},

Test{ // 16
runes.ENCODING_UTF16BE,
[]byte{0xFE,0xFF},
[]rune{0xFEFF,runes.EOF},
nil,
},

Test{ // 17
runes.ENCODING_UTF8,
[]byte{0xEF,0xBB,0xBF},
[]rune{0xFEFF,runes.EOF},
nil,
},

Test{ // 18
runes.ENCODING_8BIT,
[]byte{'a','b','c',0xC4},
[]rune{LATIN1,'a','b','c','Ä',runes.EOF},
nil,
},

Test{ // 19
runes.ENCODING_UTF8,
[]byte{0xEF,0xBB,0xBF,0xEF,0xBB},
[]rune{0xFEFF,runes.GARBAGE+0xEF,runes.GARBAGE+0xBB,runes.UNSPECIFIC_IO_ERROR},
errors.New("foo"),
},

Test{ // 20
runes.ENCODING_UTF8,
[]byte{0xEF,0xBB,0xBF,0xF0,"ü"[0],"ü"[1],0xBB,0xBF,0xBB},
[]rune{0xFEFF,runes.GARBAGE+0xF0,'ü',runes.GARBAGE+0xBB,runes.GARBAGE+0xBF,runes.GARBAGE+0xBB,runes.EOF},
nil,
},

Test{ // 21
runes.ENCODING_UTF8,
[]byte{0xEF,0xBB,0xBF,0xF0,0x90,0x8D,0x88},
[]rune{0xFEFF,'𐍈',runes.EOF},
nil,
},

Test{ // 22
runes.ENCODING_UTF8,
[]byte{0xEF,0xBB,0xBF,0xFF,0x80,0x81,0x82},
[]rune{0xFEFF,runes.GARBAGE+0xFF,runes.GARBAGE+0x80,runes.GARBAGE+0x81,runes.GARBAGE+0x82,runes.EOF},
nil,
},

Test{ // 23
runes.ENCODING_UTF8,
[]byte{0xEF,0xBB,0xBF,0xF0,0x90,0x8D},
[]rune{0xFEFF,runes.GARBAGE+0xF0,runes.GARBAGE+0x90,runes.GARBAGE+0x8D,runes.EOF},
nil,
},

Test{ // 24
runes.ENCODING_UTF8,
[]byte{0xEF,0xBB,0xBF,0xEF,0xF0,0x90,0x8D,0x88},
[]rune{0xFEFF,runes.GARBAGE+0xEF,'𐍈',runes.EOF},
nil,
},

Test{ // 25
runes.ENCODING_8BIT,
[]byte{'a','b','c',0xC4},
[]rune{ASCII,'a','b','c','�',runes.EOF},
nil,
},

Test{ // 26
runes.ENCODING_UTF8,
[]byte{0xC1, 0x80},
[]rune{runes.GARBAGE+0xC1, runes.GARBAGE+0x80, runes.EOF},
nil,
},

Test{ // 27
runes.ENCODING_UTF8,
[]byte{0xC0, 0x80},
[]rune{runes.Overlong0[0], runes.Overlong0[1], runes.EOF},
nil,
},


}



func main() {
  for i := range tests {
    fmt.Printf("Test %v\n", i+1)
    u := runes.NewUTFReader(&TestReader{&tests[i],0})
    idx := 0
    for {
      if idx == len(tests[i].Out) {
        break
      }
      r := tests[i].Out[idx]
      idx++
      if r == LATIN1 {
        u.Set8BitTable(runes.Make8BitTable("Latin-1"))
        if u.Encoding() != runes.ENCODING_8BIT {
          fmt.Printf("  Encoding() == %v, should be ENCODING_8BIT\n", u.Encoding())
          break
        }
        continue
      }
      
      if r == ASCII {
        u.Set8BitTable(runes.Make8BitTable("A-s-C-i-i"))
        if u.Encoding() != runes.ENCODING_8BIT {
          fmt.Printf("  Encoding() == %v, should be ENCODING_8BIT\n", u.Encoding())
          break
        }
        continue
      }

      r2 := u.ReadRune()

      if tests[i].Encoding != u.Encoding() {
        fmt.Printf("  Encoding() == %v, should be %v\n", u.Encoding(), tests[i].Encoding)
        break
      }
      
      if r != r2 {
        fmt.Printf("  Rune #%v is %v, should be %v\n", idx, r2, r)
        break
      }
    }
  }
}
