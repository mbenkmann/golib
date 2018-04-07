/* Copyright (C) 2016 Matthias S. Benkmann
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this file (originally named utf.go) and associated documentation files 
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

// Convert files in a variety of encodings into a sequence of runes.
package runes // import "winterdrache.de/golib/runes"

import (
         "io"
       )

// runes.Reader returns this when ReadRune() is called at the end of the stream.
const EOF = -512

// runes.Reader returns this when an error is encountered during reading
// that has no more specific error code assigned.
const UNSPECIFIC_IO_ERROR = EOF + 1

// runes.Reader returns GARBAGE+b if an undecodable byte b is encountered.
// This is only possible for multi-byte encodings that have invalid sequences.
const GARBAGE = -256

// UTFReader.Encoding() returns this before the encoding has been detected.
const ENCODING_UNKNOWN = 0

// UTFReader.Encoding() returns this after UTFReader.Set8BitTable() has been used.
// This can also be used as a bit mask to test for UTF-8 or a table encoding like
// this:
//
//   if u.Encoding() & runes.ENCODING_8BIT != 0 { /* UTF-8 or table */ }
const ENCODING_8BIT = 1

// UTFReader.Encoding() returns this when auto-detection has found UTF-8.
const ENCODING_UTF8 = 3

// Bit mask for testing for 16 bit encodings:
//
//   if u.Encoding() & runes.ENCODING_16BIT != 0 { /* UTF-16LE or UTF-16BE */ }
const ENCODING_16BIT = 4

// UTFReader.Encoding() returns this when auto-detection has found UTF-16 little-endian.
const ENCODING_UTF16LE = 4

// UTFReader.Encoding() returns this when auto-detection has found UTF-16 big-endian.
const ENCODING_UTF16BE = 12

// This sequence of error runes is returned by UTFReader when it encounters the
// overlong 2 bytes encoding of U+0000. It is provided for convenience because
// this encoding is used by some systems to avoid actual 0 bytes. If you need to
// handle this form of modified UTF-8, you can special case this sequence of
// error runes returned from UTFReader.
var Overlong0 = [2]rune{GARBAGE+0xC0, GARBAGE+0x80}

// The UTF-8 standard requires that each codepoint be encoded with the minimum
// number of bytes. MinCodePointForUTFLength[N] is the smallest codepoint that
// can validly be encoded with N bytes.
// UTF8Reader uses this array to diagnose overlong encodings.
var minCodepointForUTF8Length = [5]rune{0,0,0x80,0x800,0x10000}

// Returns true if r is a legal codepoint for UTF-8 and UTF-16.
// It will return false for error runes returned by UTFReader.
// It will also return false for non-error runes returned by
// UTFReader that represent codepoints that derive from
// technically sound UTF-8 encodings that can not be encoded
// as UTF-16 and are therefore declared illegal by the standard
// even for UTF-8.
func IsUTFLegal(r rune) bool {
  return (r >= 0 && r <= 0xD7FF) || (r >= 0xE000 && r <= 0x10FFF)
}

// A stream of runes. Errors are returned as runes as part
// of the stream instead of as out-of-band error-type returns.
type Reader interface {
  // Reads the next rune from the stream. If an
  // IO error occurs or a byte is encountered in the
  // underlying byte stream that is
  // not part of a valid encoding, it will be returned as a
  // special rune that has a value < 0 that you can evaluate
  // as follows:
  //
  //   if r < 0 {
  //     if r >= runes.GARBAGE {
  //       invalid_byte := byte(r - runes.GARBAGE)
  //     } else {
  //       switch r {
  //         case runes.EOF: ...
  //         case runes.UNSPECIFIC_IO_ERROR: ...
  //         default: error_code := r - runes.EOF
  //       }
  //     }
  //   }
  //
  // After any IO error rune has been returned, all further
  // reads will return EOF.
  ReadRune() rune
}

// A runes.Reader that converts a byte stream into a stream of runes
// with automatic detection if the input is UTF8, UTF16BE or UTF16LE.
type UTFReader struct {
  // The underlying byte stream we convert into runes.
  r io.Reader
  
  // ENCODING_UNKNOWN, ENCODING_UTF8, ENCODING_8BIT, ENCODING_UTF16LE or ENCODING_UTF16BE
  enc int
  
  // We need a max of 4 bytes to decode any valid UTF-8 character,
  // since per https://tools.ietf.org/html/rfc3629 encodings
  // that exceed the UTF-16 encodable range are invalid.
  buf [4]byte
  
  // If we discover an error in an encoding we report the first
  // byte as the error byte (or the first 2 bytes in the case of
  // UTF-16) and reparse the remaining bytes.
  // rest stores such bytes that we have already read from
  // the stream but which we need still to parse (again).
  rest []byte
  
  // If non-0, then no further reads will be performed on
  // the underlying stream and when all buffered bytes have been
  // used up, this rune will be returned from ReadRune() and then
  // eof will be changed to EOF. This variable is set when
  // an IO error is encountered so that the IO error is not
  // lost while processing previously buffered bytes.
  eof rune
  
  // If an IO error is encountered (including io.EOF)
  // it will be stored here. Can be read via IOError() method.
  ioerror error
  
  // Translation table for ENCODING_8BIT mode.
  trans *[256]rune
}

// Returns the encoding mode the UTFReader is operating in.
// This is ENCODING_UNKNOWN before the first call to ReadRune()
// or Set8BitTable() and and afterwards ENCODING_8BIT, ENCODING_UTF8,
// ENCODING_UTF16LE or ENCODING_UTF16BE.
func (u *UTFReader) Encoding() int {
  return u.enc
}

// Stops UTF decoding and instead switches to simple translation of
// bytes to runes according to table t. Note that after this call
// Encoding() will return ENCODING_8BIT. This function may be called
// at any time to switch from UTF mode (typically UTF-8) to a legacy encoding. 
func (u *UTFReader) Set8BitTable(t *[256]rune) {
  if t == nil { panic("Set8BitTable(nil) not allowed") }
  u.trans = t
  u.enc = ENCODING_8BIT
}

// Creates a translation table for use with UTFReader.Set8BitTable()
// that handles the named charset. charset will be converted to upper-case and
// all characters except [0-9A-Z] will be removed before matching the charset name,
// so "iso-8859-1" is the same as "ISO88591". If charset is not supported, returns nil.
// The following charsets are supported:
//
//   ASCII: All bytes >= 128 are mapped to the unicode replacement character U+FFFD.
//   LATIN1, ISO88591: ISO-8859-1
func Make8BitTable(charset string) *[256]rune {
  chars := make([]byte, len(charset))
  k := 0
  for i := 0; i < len(charset); i++ {
    ch := charset[i]
    if ch >= 'a' { ch = ch - 'a' + 'A' }
    if (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'Z') {
      chars[k] = ch
      k++
    }
  }
  charset = string(chars[0:k])
  var tab [256]rune
  for i := 0; i < 128; i++ { tab[i] = rune(i) } // fill in ASCII common core
  switch charset {
    case "ASCII": for i := 128; i < 256; i++ { tab[i] = rune(0xFFFD) }
    case "LATIN1", "ISO88591": for i := 128; i < 256; i++ { tab[i] = rune(i) }
    
    default: return nil
  }
  
  return &tab 
}

// Creates a new UTFReader that converts the byte stream
// r into a stream of runes. Unless Set8BitTable() is
// called before the first ReadRune(), the first ReadRune()
// will auto-detect if the input is UTF8, UTF16BE or UTF16LE.
// Set8BitTable() may be called at any time to switch from
// UTF mode (typically UTF-8) to a legacy encoding. 
//
// Each Read() from r will read the minimum
// number of bytes necessary to decode the next rune.
// You may want to wrap a bufio.Reader around r to improve
// performance if r is not buffered already.
//
// Undecodable bytes will not stop the decoding. It will continue
// with the next byte (after returning a special error rune).
// Because of the auto-syncing properties of UTF, if the input is
// at least partially correct UTF, the decoding will recover.
//
// Handling of UTF-8 errors:
//
//   * The UTF-8 standard mandates that each codepoint be encoded
//     with the minimum number of bytes. UTFReader recognizes
//     overlong encodings and returns them as error runes.
//     This guarantees that no information is lost and the original
//     byte sequence from the underlying stream can be recovered
//     from the runes returned by UTFReader.
//   * Of particular note is the overlong 2 byte encoding of
//     U+0000 which is used by some systems to avoid actual 0 bytes
//     in the encoding.
//     For your convenience the array runes.Overlong0 contains
//     the error runes returned by UTFReader when it encounters
//     this encoding, so you can special case it if desired.
//   * Some codepoints are declared illegal by the UTF-8 standard
//     for compatibility with UTF-16, which can not encode these
//     codepoints. UTFReader rejects codepoints >0x13ffff (returning
//     them as error runes) but other than that does not diagnose
//     this class of illegal codepoints. You can use the
//     function IsUTFLegal() to check if a codepoint is legal.
//   
func NewUTFReader(r io.Reader) *UTFReader {
  return &UTFReader{r:r}
}

// If UTFReader encounters an IO error (including io.EOF) while
// reading from the underlying byte stream, it can be retrieved
// with this function.
// Note that because of buffering this function may return
// non-nil before ReadRune() has returned the corresponding
// error rune. You should not use this to control read loops.
func (u *UTFReader) IOError() error {
  return u.ioerror
}

// Returns the next rune. See runes.Reader for details.
func (u *UTFReader) ReadRune() rune {
  if u.enc == ENCODING_UNKNOWN {
    n, err := io.ReadFull(u.r, u.buf[0:2])
    u.ioerror = err
    if err == io.EOF { // empty document
      u.enc = ENCODING_UTF8
      u.eof = EOF
    } else if err == io.ErrUnexpectedEOF {
      // document consisting of only 1 character
      u.enc = ENCODING_UTF8
      u.rest = u.buf[0:1]
      u.eof = EOF
    } else if err != nil {
      // IO error
      u.enc = ENCODING_UTF8
      if n == 1 {
        u.rest = u.buf[0:1]
      }
      u.eof = UNSPECIFIC_IO_ERROR
    } else {
      if u.buf[0] == 0 && u.buf[1] != 0 {
        // UTF-16BE, no BOM
        u.enc = ENCODING_UTF16BE
        u.rest = u.buf[0:2]
      } else if u.buf[1] == 0 && u.buf[0] != 0 {
        // UTF-16LE, no BOM
        u.enc = ENCODING_UTF16LE
        u.rest = u.buf[0:2]
      } else if u.buf[0] == 0xFE && u.buf[1] == 0xFF {
        // UTF-16BE, BOM
        u.enc = ENCODING_UTF16BE
        u.rest = u.buf[0:2]
      } else if u.buf[1] == 0xFE && u.buf[0] == 0xFF {
        // UTF-16LE, BOM
        u.enc = ENCODING_UTF16LE
        u.rest = u.buf[0:2]
      } else {
        // UTF-8 with or without BOM
        u.enc = ENCODING_UTF8
        u.rest = u.buf[0:2]
      }
    }
  }
  
  if u.enc <= ENCODING_UTF8 { // ENCODING_UTF8 or ENCODING_8BIT
    need := 1
    for {
      if len(u.rest) < need { // we need more bytes
        if u.eof != 0 { // we cannot read more bytes because we're at EOF already
          if len(u.rest) != 0 { // if we still have buffered bytes => return as garbage
            b := u.rest[0]
            u.rest = u.rest[1:]
            return GARBAGE + rune(b)
            
          } else {
            r := u.eof
            u.eof = EOF
            return r
          }
        }
        
        if len(u.rest) == cap(u.rest) { // if we can not append to u.rest this means we need to relocate to the start of buf
          u.rest = u.buf[0:copy(u.buf[0:], u.rest)] // This is guaranteed to have at least 1 byte additional capacity because no valid UTF8 sequence needs more than 4 bytes
        }
        
        n, err := u.r.Read(u.rest[len(u.rest):len(u.rest)+1])
        u.ioerror = err
        if err != nil {
          if err != io.EOF { 
            u.eof = UNSPECIFIC_IO_ERROR
          } else {
            u.eof = EOF
          }
        }
        
        if n == 0 {
          continue
        }
        
        u.rest = u.rest[0:len(u.rest)+1] // grow rest by 1
      }
      
      if need == 1 { // this is an individual byte or the start byte of a multi-byte sequence
        b := u.rest[0]
        
        if u.trans != nil {
          u.rest = u.rest[1:]
          return u.trans[b]
        }
        
        if b < 128 {
          u.rest = u.rest[1:]
          return rune(b)
        }
        
        if b < 0xC0 { // 0b11000000
          // an unexpected continuation byte => garbage
          u.rest = u.rest[1:]
          return GARBAGE + rune(b)
        }
        
        need++
        
        if b < 0xE0 { // 0b11100000
          continue
        }
        
        need++
        
        if b < 0xF0 { // 0b11110000
          continue
        }
        
        need++
        
        if b < 0xF5 { // 0b11110101  All codes starting with F5 (and higher) are invalid
          continue
        }
        
        // b is not a valid start byte because a sequence starting with it would lead
        // to a codepoint that is out of the range of codepoints permitted by RFC3629
        u.rest = u.rest[1:]
        return GARBAGE + rune(b)
        
      } else { // we are decoding a multi-byte sequence
      
        b := u.rest[len(u.rest)-1] // get newest byte
        if b & 0xC0 != 0x80 { // if b is not a continuation byte => return 1 byte from rest as garbage
          b = u.rest[0]
          u.rest = u.rest[1:]
          return GARBAGE + rune(b)
        }
      
        if len(u.rest) < need { // we don't have all the parts, yet => keep reading
          continue
        }
        
        // we have collected (and verified) all parts of the rune => assemble it
        r := rune((u.rest[0] << uint(need)) >> uint(need)) // clear start byte marker bits
        r = (r << 6) + rune(u.rest[1] ^ 0x80)
        if need > 2 {
          r = (r << 6) + rune(u.rest[2] ^ 0x80)
          if need > 3 {
            r = (r << 6) + rune(u.rest[3] ^ 0x80)
          }
        }
        
        if r >= minCodepointForUTF8Length[need] {
          u.rest = u.rest[need:]
          return r
        } else { // overlong encoding => return 1st byte of rest as garbage.
                 // The other bytes will also be returned as garbage on further calls because they
                 // are continuation bytes.
          b = u.rest[0]
          u.rest = u.rest[1:]
          return GARBAGE + rune(b)
        }
      }
    }
    
  } else { // UTF-16LE or UTF-16BE
    var r rune
    need := 2
    for {
      if len(u.rest)+1 < need { // +1 so that we don't enter this branch if we have an odd number of bytes in u.rest
        if u.eof != 0 { // we cannot read more bytes because we're at EOF already
          if len(u.rest) != 0 { // if we still have buffered bytes => return as garbage
            b := u.rest[0]
            u.rest = u.rest[1:]
            return GARBAGE + rune(b)
            
          } else {
            r := u.eof
            u.eof = EOF
            return r
          }
        }
        
        if len(u.rest)+2 > cap(u.rest) { // if we can not append to u.rest this means we need to relocate to the start of buf
          u.rest = u.buf[0:copy(u.buf[0:], u.rest)] 
        }
        
        n, err := io.ReadFull(u.r, u.rest[len(u.rest):len(u.rest)+2])
        u.ioerror = err
        if n == 0 {
          if err == io.EOF {
            u.eof = EOF
          } else {
            u.eof = UNSPECIFIC_IO_ERROR
          }
        } else if n == 1 {
          u.rest = u.rest[0:len(u.rest)+1] // grow rest by 1 to include the read byte
          if err == io.ErrUnexpectedEOF {
            u.eof = EOF
          } else {
            u.eof = UNSPECIFIC_IO_ERROR
          }
        } else if n == 2 {
          u.rest = u.rest[0:len(u.rest)+2] // grow rest by 2 to include the read bytes
        }
        
        continue // restart from above in case we have not read anything
      }
      
      if len(u.rest) & 1 == 1 { // we have 1 garbage byte remaining
        b := u.rest[0]
        u.rest = u.rest[1:]
        return GARBAGE + rune(b)
      }

      if need == 2 {
        if u.enc == ENCODING_UTF16BE {
          r = rune(u.rest[0])
          r = (r << 8) + rune(u.rest[1])
        } else {
          r = rune(u.rest[1])
          r = (r << 8) + rune(u.rest[0])
        }
        
        if r <= 0xD7FF || r >= 0xE000 {
          u.rest = u.rest[2:]
          return r
        }
        
        if r >= 0xDC00 && r <= 0xDFFF { // low surrogate not allowed as first => 2 garbage bytes
          b := u.rest[0]
          u.rest = u.rest[1:]
          return GARBAGE + rune(b)
        }
        
        // if we get here, r is in the range 0xD800..0xDBFF which is the high surrogate range
        // so we need a second 16bit part
        need = 4
        
      } else { // we have a verified high surrogate in r and an as yet unverified part in rest[2:4]
        highsurrogate := r
        if u.enc == ENCODING_UTF16BE {
          r = rune(u.rest[2])
          r = (r << 8) + rune(u.rest[3])
        } else {
          r = rune(u.rest[3])
          r = (r << 8) + rune(u.rest[2])
        }

        if r < 0xDC00 || r > 0xDFFF { // not a valid low surrogate => purge out garbage
          b := u.rest[0]
          u.rest = u.rest[1:]
          return GARBAGE + rune(b)
        }
        
        u.rest = u.buf[0:0]
        return 0x10000 + r - 0xDC00 + ((highsurrogate - 0xD800) << 10)
      }
    }
  }
}

