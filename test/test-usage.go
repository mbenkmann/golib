/* Written 2015 by Matthias S. Benkmann
 *
 * The author hereby waives all copyright and related rights to the contents
 * of this example file (test-usage.go) to the extent possible under the law.
 */

package main

import (
         "fmt" 
         "../argv"
       )


var usage1 = argv.Usage{
{ 0,1, "", "",argv.ArgUnknown, "1234567890ABCDEFGHIJ" },
}

var usage2 = argv.Usage{
{ 0,1, "", "",argv.ArgUnknown, "123 45 678 90AB CDE FGHI J" },
}

var usage3 = argv.Usage{
{ 0,1, "", "",argv.ArgUnknown, "一二三四五ABCDEFGHIJ" },
}

var usage4 = argv.Usage{
{ 0,1, "", "",argv.ArgUnknown, "--3456789 \t12345678901234567890" },
}

func main() {
  argv.Columns = 10
  fmt.Println(usage1)
  argv.Columns = 5
  fmt.Println(usage1)
  
  argv.Columns = 10
  fmt.Println(usage2)
  argv.Columns = 5
  fmt.Println(usage2)
  
  argv.Columns = 10
  fmt.Println(usage3)
  argv.Columns = 5
  fmt.Println(usage3)
  
  argv.Columns = 20
  argv.LastColumnMinPercent = 50
  fmt.Println(usage4)
  argv.Columns = 20
  argv.LastColumnMinPercent = 75
  argv.LastColumnOwnLineMaxPercent = 75
  fmt.Println(usage4)
  argv.Columns = 20
  argv.LastColumnMinPercent = 75
  argv.LastColumnOwnLineMaxPercent = 25
  fmt.Println(usage4)
}
