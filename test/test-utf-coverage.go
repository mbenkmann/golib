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
    
    if len(p) == 0 { return n,nil }
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

Test{ // 28
runes.ENCODING_UTF8,
[]byte{0xF0, 0x82, 0x82, 0xAC},
[]rune{runes.GARBAGE+0xF0,runes.GARBAGE+0x82,runes.GARBAGE+0x82,runes.GARBAGE+0xAC, runes.EOF},
nil,
},

Test{ // 29
runes.ENCODING_UTF8,
[]byte{0xE0, 0x82, 0xA2},
[]rune{runes.GARBAGE+0xE0,runes.GARBAGE+0x82,runes.GARBAGE+0xA2, runes.EOF},
nil,
},

Test{ // 30
runes.ENCODING_UTF16LE,
[]byte{0xFF,0xFE,0x00,0xD8},
[]rune{0xFEFF, runes.GARBAGE+0x00,runes.GARBAGE+0xD8, runes.EOF},
nil,
},

Test{ // 31
runes.ENCODING_UTF16LE,
[]byte{0xFF,0xFE},
[]rune{0xFEFF, runes.EOF},
nil,
},

Test{ // 32
runes.ENCODING_UTF16LE,
[]byte{0xFF,0xFE},
[]rune{0xFEFF, runes.UNSPECIFIC_IO_ERROR},
errors.New("foo"),
},

Test{ // 33
runes.ENCODING_UTF16LE,
[]byte{0xFF,0xFE,0xD8},
[]rune{0xFEFF, runes.GARBAGE+0xD8, runes.EOF},
nil,
},

Test{ // 34
runes.ENCODING_UTF16BE,
[]byte{0xFE,0xFF,0xD8,0x52,0xDF,0x62, 0xD8,0x52,0xDF,0x62},
[]rune{0xFEFF, 0x24B62, 0x24B62, runes.EOF},
nil,
},

Test{ // 35
runes.ENCODING_UTF16BE,
[]byte{0xFE,0xFF,0xD8,0x01,0xDC,0x37, 0xD8,0x01,0xDC,0x37},
[]rune{0xFEFF, 0x10437, 0x10437, runes.EOF},
nil,
},

Test{ // 36
runes.ENCODING_UTF16BE,
[]byte{0xFE,0xFF,0x12,0x34,0xED,0xCB, 0x43,0x21,0xEF,0xFC},
[]rune{0xFEFF,0x1234,0xEDCB,0x4321,0xEFFC, runes.EOF},
nil,
},

Test{ // 37
runes.ENCODING_UTF16BE,
[]byte{0xFE,0xFF,0xDC,0x00,0xDF,0xFF},
[]rune{0xFEFF,runes.GARBAGE+0xDC,runes.GARBAGE+0x00,runes.GARBAGE+0xDF,runes.GARBAGE+0xFF, runes.EOF},
nil,
},

Test{ // 38
runes.ENCODING_UTF16BE,
[]byte{0xFE,0xFF,0xD8,0x00,0xD8,0x00},
[]rune{0xFEFF,runes.GARBAGE+0xD8,runes.GARBAGE+0x00,runes.GARBAGE+0xD8,runes.GARBAGE+0x00, runes.EOF},
nil,
},

Test{ // 39
runes.ENCODING_UTF16BE,
[]byte{0xFE,0xFF,0xD8,0x00},
[]rune{0xFEFF, runes.GARBAGE+0xD8,runes.GARBAGE+0x00, runes.EOF},
nil,
},

Test{ // 40
runes.ENCODING_UTF16BE,
[]byte{0xFE,0xFF},
[]rune{0xFEFF, runes.EOF},
nil,
},

Test{ // 41
runes.ENCODING_UTF16BE,
[]byte{0xFE,0xFF},
[]rune{0xFEFF, runes.UNSPECIFIC_IO_ERROR},
errors.New("foo"),
},

Test{ // 42
runes.ENCODING_UTF16BE,
[]byte{0xFE,0xFF,0xD8},
[]rune{0xFEFF, runes.GARBAGE+0xD8, runes.EOF},
nil,
},

Test{ // 43
runes.ENCODING_UTF16LE,
[]byte{0xFF,0xFE,0x52,0xD8,0x62,0xDF, 0x52,0xD8,0x62,0xDF},
[]rune{0xFEFF, 0x24B62, 0x24B62, runes.EOF},
nil,
},

Test{ // 44
runes.ENCODING_UTF16LE,
[]byte{0xFF,0xFE,0x01,0xD8,0x37,0xDC, 0x01,0xD8,0x37,0xDC},
[]rune{0xFEFF, 0x10437, 0x10437, runes.EOF},
nil,
},

Test{ // 45
runes.ENCODING_UTF16LE,
[]byte{0xFF,0xFE,0x34,0x12,0xCB,0xED, 0x21,0x43,0xFC,0xEF},
[]rune{0xFEFF,0x1234,0xEDCB,0x4321,0xEFFC, runes.EOF},
nil,
},

Test{ // 46
runes.ENCODING_UTF16LE,
[]byte{0xFF,0xFE,0x00,0xDC,0xFF,0xDF},
[]rune{0xFEFF,runes.GARBAGE+0x00,runes.GARBAGE+0xDC,runes.GARBAGE+0xFF,runes.GARBAGE+0xDF, runes.EOF},
nil,
},

Test{ // 47
runes.ENCODING_UTF16LE,
[]byte{0xFF,0xFE,0x00,0xD8,0x00,0xD8},
[]rune{0xFEFF,runes.GARBAGE+0x00,runes.GARBAGE+0xD8,runes.GARBAGE+0x00,runes.GARBAGE+0xD8, runes.EOF},
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
  
  for i, cp := range []rune{0,-1,0xd7ff,0xd800,0xe000,0xdfff,0x10fff,0x11000} {
    if runes.IsUTFLegal(cp) != ((i&1) == 0) {
      fmt.Printf("Code point %x incorrectly categorized by IsUTFLegal()\n", cp)
    }
  }
  
}
