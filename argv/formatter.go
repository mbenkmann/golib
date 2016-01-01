/* Copyright (C) 2015 Matthias S. Benkmann
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this file (originally named formatter.go) and associated documentation files 
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


package argv

import (
         "os"
         "fmt"
         "strconv"
       )

/*
 A multi-level iterator over the OptionInfo.Help texts of a Usage.
 The general idea is that LinePartIterator will return substrings
 of the Help texts in the order in which they need to be written
 to the screen to achieve a table layout as described in String.Usage().
 The formatting aspects LinePartIterator is not concerned with are
 line wrapping and inserting spaces to achieve column alignment.
 
 The top-level organizational unit is the TABLE.
 
 A table begins at an OptionInfo with Help != "\f" and extends up to
 a Descriptor with Help == "\f".
 
 A table consists of ROWS.
 
 Each OptionInfo.Help (except for the special string "\f" which is a table
 separator) starts a new row. Every '\n' character inside Help also starts
 a new row, even if it is the last character.
 This means that the number of rows defined by an OptionInfo.Help string is
 equal to the number of '\n' contained in the string PLUS ONE.
 
 A row consists of CELLS.
 
 Cells within a row are separated by '\t'.
 
 The word COLUMN refers to the index of a cell within its row.
 E.g. column 0 contains the very first cell of a row.
 If that sounds confusing simply consider "column" as a synonym for "cell".
 
 A cell consists of PARTS.
 
 Parts within a cell are separated by '\v'. Usage.String() stacks parts
 vertically, so this is essentially a line break within a cell.
 ('\n' would start a new row in the table which is NOT equivalent unless
 your table has only 1 row)
 
 Rows within a table need not have the same number of cells and cells
 need not have the same number of parts.
 
 LinePartIterator iterates through the Help texts at 3 levels: table, row and part.
 Tables and rows are visited in their natural order. Parts are visited in the
 following way:
 Within the current row, it will visit the 1st part of the 1st cell, then the
 1st part of the 2nd cell, and so on until the 1st part of the last cell of
 the current row has been visited. Then the 2nd part of the 1st cell is visited,
 followed by the 2nd part of the 2nd cell, and so on.
 
 Example: The row "1 \v 3 \t 2 \v 4" has 2 cells and 4 parts.
 The parts will be returned in the order 1, 2, 3, 4.
 
 It is possible that some cells have fewer parts than others. In this case
 LinePartIterator will "fill up" these cells with 0-length parts. IOW, LinePartIterator
 always returns the same number of parts for each cell within the same row (but not for
 cells in different rows).
 */
type LinePartIterator struct {
  usage Usage // the Usage being iterated over
  tablestart int // usage[tablestart] is the 1st OptionInfo of the current table.
  rowdesc int // usage[rowdesc] is the OptionInfo that contains the current row.
  rowstart int // index of 1st character of current row within usage[rowdesc].Help.
  ptr int // index of 1st character of current part within usage[rowdesc].Help.
  col int // Index of current column.
  length int // Length of the current part (that ptr points at) in BYTES
  screenlen int // Length of the current part in screen columns (taking narrow/wide chars into account).
  max_subrow int // Greatest index of a line within the block. This is the number of \v within the cell with the most \vs.
  subrow int // Line index of the current part within the current cell.
  target_subrow int // Line index of the parts we should return to the user on this iteration.
  hit_subrow bool // Flag whether we encountered a part with line index target_subrow in the current cell.
}

// Creates an iterator for usage.
func (usage Usage) Iterate() *LinePartIterator {
  return &LinePartIterator{usage:usage, tablestart:0, rowdesc:-1, rowstart:0, ptr:-1, col:-1, length:0, screenlen:0,
  max_subrow:0, subrow:0, target_subrow:0, hit_subrow:true}
}

/*
 Moves iteration to the next table (if any). Has to be called once on a new
 LinePartIterator to move to the 1st table.
 Returns false if moving to next table failed because no further table exists.
 E.g.
   
   for iter := usage.Iterate(); iter.NextTable() {
     ...
   }
*/
func (l *LinePartIterator) NextTable() bool {
  // If this is NOT the first time NextTable() is called after the constructor,
  // then skip to the next table break (i.e. an OptionInfo with help == "\f")
  if l.rowdesc != -1 {
    for l.tablestart < len(l.usage) && l.usage[l.tablestart].Help != "\f" {
      l.tablestart++
    }
  }

  // Find the next table after the break (if any)
  for l.tablestart < len(l.usage) && l.usage[l.tablestart].Help == "\f" {
    l.tablestart++
  }

  return l.RestartTable()
}

/*
 Reset iteration to the beginning of the current table.
 Returns false if there is no current table (because the iteration has
 finished).
 After calling this function, you must call NextRow() to move to the 1st
 row.
*/
func (l *LinePartIterator) RestartTable() bool {
  l.rowdesc = l.tablestart
  l.rowstart = 0
  l.ptr = -1;
  return l.tablestart < len(l.usage)
}

/*
  Moves iteration to the next row (if any). Has to be called once after each call to
  NextTable() to move to the 1st row of the table.
  Returns false if moving to next row failed because no further row exists within
  the current table. Note that NextRow() will not enter the next table. You need
  to use NextTable() for that.
*/
func (l *LinePartIterator) NextRow() bool {
  if l.rowdesc == -1 || l.rowdesc >= len(l.usage) || l.usage[l.rowdesc].Help == "\f" { // end of Usage or table break
    return false
  }
  
  if l.ptr == -1 {
    return l.RestartRow();
  }

  for l.ptr < len(l.usage[l.rowdesc].Help) && l.usage[l.rowdesc].Help[l.ptr] != '\n' {
    l.ptr++
  }

  if l.ptr == len(l.usage[l.rowdesc].Help) {
    l.rowdesc++
    l.rowstart = 0
    
  } else { // if l.usage[l.rowdesc].Help[l.ptr] == '\n'
    l.rowstart = l.ptr + 1;
  }

  return l.RestartRow()
}

/*
 Reset iteration to the beginning of the current row.
 Returns false if there is no current row (because iteration has
 already passed the last row of the current table).
*/
func (l *LinePartIterator) RestartRow() bool {
  if l.rowdesc == -1 { return false }
  l.ptr = l.rowstart;
  l.col = -1;
  l.length = 0;
  l.screenlen = 0;
  l.max_subrow = 0;
  l.subrow = 0;
  l.target_subrow = 0;
  l.hit_subrow = true;
  return l.rowdesc < len(l.usage) && l.usage[l.rowdesc].Help != "\f"
}

/**
 * Moves iteration to the next part (if any). Has to be called once after each call to
 * NextRow() to move to the 1st part of the row.
 * Returns false if moving to next part failed because no further part exists.
 *
 * See LinePartIterator for details about the iteration.
 */
func (l *LinePartIterator) NextPart() bool {
  // return false if called directly after construction, NextTable() or RestartTable()
  // without calling NextRow() first
  if (l.ptr == -1) {
    return false
  }
  
  // Return false if row pointer is at end of Usage or table break
  if l.rowdesc >= len(l.usage) || l.usage[l.rowdesc].Help == "\f" {
    return false
  }

  // if this is the first call after NextRow() or RestartRow()
  if (l.col == -1) {
    l.col = 0
    l.update_length()
    return true
  }

  l.ptr += l.length // go to 1st character following current part
  
  for {
    ch := byte(0) // if we're at the end of the Help string
    if l.ptr < len(l.usage[l.rowdesc].Help) {
      ch = l.usage[l.rowdesc].Help[l.ptr]
    }
    
    switch (ch) {
      case '\v':
        l.subrow++
        upmax(&l.max_subrow, l.subrow)
        l.ptr++
      case '\t':
        if !l.hit_subrow { // if previous column did not have the targetline
          // then "insert" a 0-length part
          l.update_length()
          l.hit_subrow = true
          return true
        }

        l.hit_subrow = false
        l.subrow = 0
        l.col++
        l.ptr++
      case 0, '\n':
        if !l.hit_subrow { // if previous column did not have the targetline
          // then "insert" a 0-length part
          l.update_length()
          l.hit_subrow = true
          return true
        }

        l.target_subrow++
        if l.target_subrow > l.max_subrow {
          l.update_length()
          return false
        }

        l.hit_subrow = false
        l.subrow = 0;
        l.col = 0;
        l.ptr = l.rowstart
        continue
      default:
        l.ptr++
        continue
    } // switch

    if l.subrow == l.target_subrow {
      l.update_length()
      l.hit_subrow = true
      return true
    }
  } // for
}

/*
 Returns the index (counting from 0) of the column in which
 the part returned by Data() is located.
*/
func (l *LinePartIterator) Column() int {
  return l.col
}

/*
 Returns the index (counting from 0) of the current part within the current
 cell. When iterating over a row, indexes change like this:
 Column()==0/Subrow()==0, Column()==1/Subrow()==0, Column()==2/Subrow()==0,...,
 Column()==0/Subrow()==1, Column()==1/Subrow()==1, Column()==2/Subrow()==1,...,
 Column()==0/Subrow()==2, Column()==1/Subrow()==2, Column()==2/Subrow()==2,...
*/
func (l *LinePartIterator) Subrow() int {
  return l.target_subrow; // NOT subrow !!! It would be wrong if !hit_subrow
}

/*
 Returns the width in screen columns of the part returned by Data().
 Takes multi-byte UTF-8 sequences and wide characters into account.
*/
func (l *LinePartIterator) ScreenLength() int {
  return l.screenlen
}

/*
  Returns the current part of the iteration.
*/
func (l *LinePartIterator) Data() string {
  if l.rowdesc == -1 || l.ptr == -1 || l.rowdesc >= len(l.usage) { return "" }
  return l.usage[l.rowdesc].Help[l.ptr:l.ptr+l.length]
}

/* 
 * Determines the byte and character lengths of the part at ptr and 
 * stores them in length and screenlen respectively.
 */
func (l *LinePartIterator) update_length() {
  l.screenlen = 0
  l.length = 0
  help := l.usage[l.rowdesc].Help
  ptr := l.ptr
  for ptr < len(help) && help[ptr] != '\v' && help[ptr] != '\t' && help[ptr] != '\n' {
    l.length++
    l.screenlen++
    startbyte := help[ptr]
    ptr++
    if (startbyte > 0xC1) { // everything <= 0xC1 (yes, even 0xC1 itself) is not a valid UTF-8 start byte
      ch := uint(clear_utf8_len(startbyte)) // mask out length bits, we don't verify their correctness
      for ptr < len(help) && help[ptr] ^ 0x80 <= 0x3F { // while next byte is continuation byte
        ch = (ch << 6) ^ uint(help[ptr] ^ 0x80) // add continuation to char code
        ptr++
        l.length++
      }
      // ch is the decoded unicode code point
      if (ch >= 0x1100 && isWideChar(ch)) { // the test for 0x1100 is here to avoid the function call in the Latin case
        l.screenlen++
      }
    }
  }
}

func formatUsage(usage Usage) string {
  columns := Columns
  if columns <= 0 {
    var err error
    columns, err = strconv.Atoi(os.Getenv("COLUMNS"))
    if err != nil { columns = 80 }
  }
  
  return fmt.Sprintf("%v", columns)
}


// Returns b with leading 1 bits (which determine the length of a
// UTF-8 byte sequence) cleared.
func clear_utf8_len(b byte) byte {
  b2 := b ^ 128
  if b2 > b { return b }
  b = b2 ^ 64
  if b > b2 { return b2 }
  b2 = b ^ 32
  if b2 > b { return b }
  b = b2 ^ 16
  if b > b2 { return b2 }
  b2 = b ^ 8
  if b2 > b { return b }
  b = b2 ^ 4
  if b > b2 { return b2 }
  return b
}

/*
 Sets *i1 = max(i1, i2)
*/
func upmax(i1 *int, i2 int) {
  if i2 > *i1 {
    *i1 = i2
  }
}

/*
 Returns true if ch is the unicode code point of a wide character.

 The following character ranges are treated as wide

 1100..115F
 2329..232A  (just 2 characters!)
 2E80..A4C6  except for 303F
 A960..A97C
 AC00..D7FB
 F900..FAFF
 FE10..FE6B
 FF01..FF60
 FFE0..FFE6
 1B000......
*/
func isWideChar(ch uint) bool {
  if (ch == 0x303F) {
    return false;
  }

  return ((0x1100 <= ch && ch <= 0x115F) || (0x2329 <= ch && ch <= 0x232A) || (0x2E80 <= ch && ch <= 0xA4C6) ||
          (0xA960 <= ch && ch <= 0xA97C) || (0xAC00 <= ch && ch <= 0xD7FB) || (0xF900 <= ch && ch <= 0xFAFF) ||
          (0xFE10 <= ch && ch <= 0xFE6B) || (0xFF01 <= ch && ch <= 0xFF60) || (0xFFE0 <= ch && ch <= 0xFFE6) ||
          (0x1B000 <= ch));
}
                   