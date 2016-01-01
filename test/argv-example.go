/* Written 2015 by Matthias S. Benkmann
 *
 * The author hereby waives all copyright and related rights to the contents
 * of this example file (argv-example.go) to the extent possible under the law.
 */

package main

import (
         "os"
         "fmt" 
         "../argv"
       )


const (
  UNKNOWN = iota
  HELP
  OPTIONAL
  REQUIRED
  FOO
  BAR
  NUMERIC
  VERBOSE
  NONEMPTY
)

const DISABLED = 0
const ENABLED = 1

var usage = argv.Usage{
{ UNKNOWN, 1, "", "",        argv.ArgUnknown, "USAGE: example_arg [options]\n\n" +
                                              "Options:" },
{ HELP,    1, "", "help",    argv.ArgNone,    "  \t--help  \tPrint usage and exit." },
{ OPTIONAL,1, "o","optional",argv.ArgOptional,"  -o[<arg>], \t--optional[=<arg>]" +
                                              "  \tTakes an argument but is happy without one." },
{ REQUIRED,1, "r","required",argv.ArgRequired,"  -r <arg>, \t--required=<arg>  \tMust have an argument." },
{ FOO,ENABLED,"f","enable-foo" ,argv.ArgNone, "  -f, \t--enable-foo  \tEnables switch foo." },
{ FOO,DISABLED,"","disable-foo",argv.ArgNone, "  \t--disable-foo  \tOverrides an --enabled-foo earlier on the command line." },
{ BAR,ENABLED,"b","enable-bar" ,argv.ArgNone, "  \t--enable-bar  \tEnables switch bar." },
{ NUMERIC, 1, "n","numeric", argv.ArgInt,     "  -n <num>, \t--numeric=<num>  \tRequires a number as argument." },
{ VERBOSE, 1, "v","verbose", argv.ArgNone,    "  -v, \t--verbose  \tIncrease verbosity." },
{ NONEMPTY,1, "1","nonempty",argv.ArgNonEmpty,"  -1 <arg>, \t--nonempty=<arg>" +
                                              "  \tCan NOT take the empty string as argument." },
{ UNKNOWN, 1, "", "",        argv.ArgUnknown,
`
Examples:
  argv-example --unknown -o -n10
  argv-example -o -n10 file1 file2
  argv-example -nfoo file1 file2
  argv-example --optional -- file1 file2
  argv-example --optional file1 file2
  argv-example --optional=file1 file2
  argv-example --optional=  file1 file2
  argv-example -o file1 file2
  argv-example -ofile1 file2
  argv-example -unk file1 file2
  argv-example -r -- file1
  argv-example -r file1
  argv-example --required
  argv-example --required=file1
  argv-example --nonempty= file1
  argv-example --nonempty=foo --numeric=999 --optional=bla file1
  argv-example --enable-foo --disable-foo -f --disable-foo
  argv-example --disable-foo -f
  argv-example -1foo
  argv-example -v -v -vv --verbose -verbose -verb
  argv-example -1 --
  argv-example -1 ""
  argv-example -fvfv
  argv-example -nu=10
  argv-example --nu=10
  argv-example -num=10
  argv-example -f - -v file2
  argv-example --enable-f --enable-b
  argv-example --enable
  argv-example -cool
` },
}

func main() {
  options, nonoptions, err, alloptions := argv.Parse(os.Args[1:], usage, "gnu -perl --abb")
  if err != nil {
    fmt.Fprintf(os.Stderr, "%v\n", err)
    os.Exit(1)
  }

  if (options[HELP].Is(ENABLED) || len(os.Args) == 1) {
    fmt.Fprintf(os.Stdout, "%v\n", usage)
    os.Exit(0)
  }

  for i, opt := range alloptions {
    fmt.Fprintf(os.Stdout, "Option #%d is ", i+1)
    switch opt.Id() {
      case HELP:
        // not possible, because handled further above and exits the program
      case OPTIONAL:
        if opt.HasArg {
          fmt.Fprintf(os.Stdout, "--optional with optional argument '%s'\n", opt.Arg)
        } else {
          fmt.Fprintf(os.Stdout, "--optional without the optional argument\n")
        }
      case REQUIRED:
        fmt.Fprintf(os.Stdout, "--required with argument '%s'\n", opt.Arg)
      case NUMERIC:
        fmt.Fprintf(os.Stdout, "--numeric with argument %d\n", opt.Value)
      case NONEMPTY:
        fmt.Fprintf(os.Stdout, "--nonempty with argument '%s'\n", opt.Arg)
      case FOO:
        fmt.Fprintf(os.Stdout, "%v (final state is ", opt.Name)
        if opt.Is(ENABLED) {
          fmt.Fprintf(os.Stdout, "--enable-foo)\n")
        } else {
          fmt.Fprintf(os.Stdout, "--disable-foo)\n")
        }
      case BAR:
        fmt.Fprintf(os.Stdout, "--enable-bar\n")
      case VERBOSE:
        fmt.Fprintf(os.Stdout, "--verbose\n")
      case UNKNOWN:
        // not possible because ArgUnknown returns ARG_ILLEGAL
        // which aborts the parse with an error
    }
  }
  
  fmt.Fprintf(os.Stdout, "verbosity level: %v\n", options[VERBOSE].Count())

  for i, opt := range nonoptions {
    fmt.Fprintf(os.Stdout, "Non-option argument #%d is %s\n", i+1, opt)
  }
}
