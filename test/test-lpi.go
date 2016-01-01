/* Written 2015 by Matthias S. Benkmann
 *
 * The author hereby waives all copyright and related rights to the contents
 * of this example file (test-lpi.go) to the extent possible under the law.
 */

package main

import (
         "os"
         "fmt" 
         "../argv"
       )

var usage = argv.Usage{
{ 0,0,"","",argv.ArgNone,"ö\vb\tc\vd\ve\tf\vß\vh\vi\nj\vk\vl\vm\tn\vo\vp\tq\vr\n" },
{ 0,0,"","",argv.ArgNone,"1\v2\v3\t\t4\v5\n" },
{ 0,0,"","",argv.ArgNone,"" },
{ 0,0,"","",argv.ArgNone,"xyz" },
{ 0,0,"","",argv.ArgNone,"\f" },
{ 0,0,"","",argv.ArgNone,"abcäöüｶﾀｶﾅ\t漢字12" },
{ 0,0,"","",argv.ArgNone,"" },
{ 0,0,"","",argv.ArgNone,"\f" },
{ 0,0,"","",argv.ArgNone,"" },
{ 0,0,"","",argv.ArgNone,"" },
}

var expected = []string{
"ö,c,f",
"b,d,ß",
"_,e,h",
"_,_,i",
"---",
"j,n,q",
"k,o,r",
"l,p,_",
"m,_,_",
"---",
"_",
"---",
"1,_,4",
"2,_,5",
"3,_,_",
"---",
"_",
"---",
"_",
"---",
"xyz",
"---",
"===",
"abcäöüｶﾀｶﾅ,漢字12",
"---",
"_",
"---",
"===",
"_",
"---",
"_",
"---",
"===",
}

func screenLength(s string) int {
  l := 0
  for _, r := range s {
    if r <= 256 || r == 'ｶ' || r == 'ﾀ' || r == 'ﾅ' { l++ } else { l+=2 }
  }
  return l
}

func main() {
  fmt.Printf("Testing if calling functions in incorrect order does not crash program...")
  usage.Iterate().NextPart()
  usage.Iterate().Data()
  usage.Iterate().NextRow()
  usage.Iterate().RestartRow()
  usage.Iterate().RestartTable()
  iter := usage.Iterate()
  iter.NextPart()
  iter.NextRow()
  iter = usage.Iterate()
  iter.NextRow()
  iter.NextPart()
  iter = usage.Iterate()
  iter.NextTable()
  iter.NextPart()
  iter = usage.Iterate()
  iter.NextTable()
  iter.NextPart()
  fmt.Printf("OK\n")
  
  current := ""
  data := []string{}
  for iter = usage.Iterate(); iter.NextTable(); {
    tableparts := []string{}
    for iter.NextRow() {
      rowparts := []string{}
      
      current = ""
      col := 0
      subrow := -1
      for iter.NextPart() {
        if iter.Column() == 0 {
          subrow++
          col = 0
        }
      
        if col != iter.Column() || subrow != iter.Subrow() {
          fmt.Printf("Column()/Subrow() is reported as, but should be %v/%v\n", iter.Column(), iter.Subrow(), col, subrow)
          os.Exit(1)
        }
        
        col++
        
        if iter.ScreenLength() != screenLength(iter.Data()) {
          fmt.Printf("Part \"%v\" is reported with ScreenLength()==%v, but should be %v\n", iter.Data(), iter.ScreenLength(), screenLength(iter.Data()))
          os.Exit(1)
        }
        
        if iter.Column() == 0 && iter.Subrow() > 0 {
          data = append(data, current)
          current = ""
        }
        if current != "" { current += "," }
        current += iter.Data()
        if iter.Data() == "" { current += "_" }
        
        rowparts = append(rowparts,iter.Data())
        tableparts = append(tableparts,iter.Data())
      }
      if current != "" {
        data = append(data, current)
      }
        
      data = append(data, "---")
      
      fmt.Printf("RestartRow()...")
      iter.RestartRow()
      for i:=0; iter.NextPart(); i++ {
        if iter.Data() != rowparts[i] {
          fmt.Println("FAILED")
          os.Exit(1)
        }
      }
      fmt.Println("OK")
      
    }
    
    if iter.NextRow() {
      fmt.Printf("NextRow() returns true after end of table\n")
      os.Exit(1)
    }
    
    if iter.RestartRow() {
      fmt.Printf("RestartRow() returns true after end of table\n")
      os.Exit(1)
    }
    
    data = append(data, "===")
    
    fmt.Printf("RestartTable()...")
    iter.RestartTable()
    i := 0
    for iter.NextRow() {
      for iter.NextPart() {
        if iter.Data() != tableparts[i] {
          fmt.Printf("FAILED (#%v: %v != %v)\n", i, iter.Data(), tableparts[i])
          os.Exit(1)
        }
        i++
      }
    }
    fmt.Println("OK")
  }
  
  for i, d := range data {
    fmt.Printf("%v\t\t\t%v ... ",d,expected[i])
    if d == expected[i] { fmt.Println("OK") } else { 
      fmt.Println("FAIL")
      os.Exit(1)
    }
  }
}
