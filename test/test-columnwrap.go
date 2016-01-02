/* Written 2015 by Matthias S. Benkmann
 *
 * The author hereby waives all copyright and related rights to the contents
 * of this example file (test-columnwrap.go) to the extent possible under the law.
 */

package main

import (
         "os"
         "fmt" 
         "../argv"
       )

var tab =  [][]string{
              []string{"Column 1,1 ", "Column 2,1 ", "This is an overlong text that doesn't fit."},
              []string{"Column 1,2 ", "Column 2,2 ", ""},
              nil,
              []string{"Column 1,3 ", "Column 2,3 ", "Hello, 世界. This is another stupid very long text."},
           }

func main() {
  x1 := len(tab[0][0])+len(tab[0][1])
  w := argv.NewColumnWrapper(x1, x1+7)
  for i := 0; i < len(tab); i++ {
    if tab[i] == nil {
      w.Flush(os.Stdout)
    } else {
      for column := 0; column < len(tab[i]); column++ {
        if column < len(tab[i])-1 {
          fmt.Fprintf(os.Stdout, tab[i][column])
        } else {
          w.Process(os.Stdout, tab[i][column])
        }
      }
      fmt.Fprintf(os.Stdout, "\n")
    }
  }
  w.Flush(os.Stdout)
}
