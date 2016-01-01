/* Copyright (C) 2015 Matthias S. Benkmann
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this file (originally named optionparser.go) and associated documentation files 
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
         "fmt"
         "strings"
         "strconv"
       )

// Number of screen columns for formatting usage in function Usage.String().
// If Columns I<= 0, then the environment variable COLUMNS is used. If COLUMNS
// is empty or cannot be parsed as an integer, 80 is used.
var Columns = 0

var ARG_OK error = nil
type ARG_NONE struct { error }
func ARG_NONE_Err(optionname string) ARG_NONE { return  ARG_NONE{fmt.Errorf("Option %v takes no argument", optionname)} }


/*
  A function that Parse() calls to check if opt.HasArg and opt.Arg
  are valid for the kind of option the ArgChecker is for.
  The return value is one of the following:
  
    - ARG_OK (which is nil but should be used for readability) if
      HasArg==true and Arg is valid; or if HasArg==false and the
      option does not take an argument.
      
    - An error of type ARG_NONE (Tip: use the ARG_NONE_Err() function) 
      if the option does not take an argument or if the argument that
      is provided is not acceptable but that's okay because the argument
      is optional.
      This may be returned regardless of the values of HasArg and Arg.
      In situations where the interpretation of a word in the argument
      vector is ambiguous (e.g. "-f -g" where "-g" could be e.g. a filename
      argument for the "-f" option or could be a "-g" option) , the Parse()
      function will first try to interpret it as an option argument passing
      opt.HasArg==true to the ArgChecker. If the ArgChecker returns ARG_NONE,
      Parse() will set HasArg to false and try to interpret the word as
      an option instead (which may result in an error if it is not a valid option).
      If the situation is unambiguous (e.g. "--foo=-g") and ARG_NONE is returned,
      the ARG_NONE error will abort Parse() with that ARG_NONE error being
      returned.
      If HasArg==false and ArgChecker returns an ARG_NONE error, it will
      be ignored (i.e. treated as if it was ARG_OK).
    
    - Any other kind of error will be returned from Parse() immediately.
*/
type ArgChecker func(opt *Option) error

/*
  Describes a command line option. OptionInfos provide the information
  for Parse() to parse an argument vector.
*/
type OptionInfo struct {
  /* An identifier. Most options have a unique identifier, but related
  options may have the same identifier with different values for
  State (see below). The Parse() function creates one entry in the
  returned options slice for each non-negative Id, so you should
  use consecutive non-negative numbers as Id. */
  Id int
  
  /* Used to differentiate related options with the same Id value.
  The typical example would be --enable-foo and --disable-foo which
  would both use the same Id==FOO but have State==ENABLED vs
  State==DISABLED.
  */
  State int
  
  /*
    Each character in this string is accepted as a short option character.
    E.g.
    if this string is "abc", then "-a", "-b" and "-c" will be accepted.
    If Short is "", there are no short options.
    NOTE: Don't include the "-" character here or stuff will break!
  */
  Short string
  
  /*
    The long option name without any leading "-".
    E.g. if Long is "foo", then "--foo" (and "-foo" if single-minus long options
    are enabled) will be accepted.
    If Long is "", there is no long option name.
  */
  Long string
  
  /*
    Checks a potential argument (or lack thereof) for validity with respect to
    the option described by this OptionInfo.
  */
  CheckArg ArgChecker
  
  /*
    String for use by Usage.String(). See there for formatting details.
  */
  Help string
}

/*
  The set of all possible command line options. The order of entries is important
  only in case of overlapping definitions (e.g. 2 options with the same long name)
  which you should avoid, and for Usage.String() which processes the Help strings
  in order.
*/
type Usage []OptionInfo

/*
  Parse() parses an argument vector into Option structures.
*/
type Option struct {
  /* 
    Useful mostly for inclusion in error messages, this is the name of the
    option as used in the actual argument vector, including leading "-" character(s).
    If the option is part of a single-letter option group in the actual argument vector,
    Name will still only contain the name of the one option represented by
    the Option object (e.g. Name=="-b" even if the option group is "-abc").
  */
  Name string
  
  /*
    If the option has an argument (e.g. in "-f filename" the argument would be filename)
    this is true.
  */
  HasArg bool
  
  /*
    true if the option has an argument that is part of the same word in the 
    argument vector (e.g. "--numeric=10" or "-n10", but not "-n 10").
  */
  ArgAttached bool
  
  /*
    If HasArg == true, this is the option's argument. Otherwise this is "".
    Note that "" is a valid argument, so you must check HasArg.
  */
  Arg string
  
  /*
    The element of the Usage that defines this option.
  */
  Info *OptionInfo
  
  /*
    The ArgChecker function used to validate Arg may store information in this
    field, typically a cooked form of the argument. E.g. ArgInt stores the argument
    converted to int in this field.
  */
  Value interface{}
  
  /*
    Points to the next Option with the same Info.Id. If isNotLast==false, this
    points to the FIRST Option with the same Info.Id in the argument vector.
  */
  next *Option
  isNotLast bool
}

// Returns the name of the option, so that Option objects can be used with *printf().
func (o *Option) String() string { return o.Name }

/*
  Returns the last option in the argument vector with the same Info.Id as o.
  When multiple contradictory options are used in the same argument vector,
  it is customary that the last one wins. E.g. if you have options --quiet
  and --verbose you would use the same OptionInfo.Id for both options
  and starting from any of them Last() would give you the option that
  actually matters. Most of the time, however, you would not use Last()
  directly but instead use the Is() method.
*/
func (o *Option) Last() *Option { 
  if o == nil { return nil }
  for o.isNotLast { o = o.next }
  return o
}

/*
  If the option in the argument vector that produced o was followed by
  another option with the same OptionInfo.Id, then Next() returns a
  pointer to the parsed Option for that other option. 
  Otherwise this function returns nil. Also returns nil if o is undefined.
*/
func (o *Option) Next() *Option {
  if o == nil || o.next == nil { return nil }
  return o.next
}

/*
  Returns the number of times options with the same Id() as this option
  occur in the argument vector.
  Use this to implement cumulative options, such as -v, -vv, -vvv for
  different verbosity levels.
  If o == nil or *o is uninitialized, returns 0.
*/
func (o *Option) Count() int {
  if o == nil || o.next == nil { return 0 }
  o = o.Last().next
  count := 1
  for o.isNotLast {
    o = o.next
    count++
  }
  return count
}

/*
  This function is the most common way to test if an option is set.
  It returns true iff o is non-nil and o.Last().Info.State == value.
  If called on an undefined option (i.e. nil), it will return false
  no matter what value is passed. For readability purposes, however,
  it is usually best to write something like option.Is(ENABLED) even
  for options that can only be present or absent in the argument vector.
  A common use case for this function are option pairs like 
  "--enable-foo" and "--disable-foo". In the Usage you would define
  the respective OptionInfos with the same Id==FOO but State==ENABLED
  vs State==DISABLED. With these definitions in place you can write
  tests like option.Is(ENABLED). This assumes that disabled is the
  default state. Because Is() returns false if none of the options
  from the pair is present in the argument vector, if the default state
  is enabled, you need to write the test as option.Is(DISABLED).
*/
func (o *Option) Is(value int) bool {
  if o == nil { return false }
  o = o.Last()
  if o.Info == nil { return false }
  return value == o.Info.State
}


/*
  Returns the OptionInfo.Id of o or -1 if o is nil or *o is uninitialized. You should
  use this function instead of accessing o.Info.Id directly because
  it handles nil.
*/
func (o *Option) Id() int {
  if o == nil || o.Info == nil { return -1 }
  return o.Info.Id
}

/*
  ArgChecker for unknown options.
  Usage should contain at least one OptionInfo with Long==Short=="" and
  ArgChecker ArgUnknown. When Parse() encounters a word on the argument vector
  that starts with a "-", is not "-" or "--", and does not match any known
  option, Parse() will call the ArgChecker for the first dummy entry in Usage
  (a dummy entry is one with Long==Short==""). If you use ArgUnknown for this
  dummy entry, Parse() will abort. If there is no dummy entry in usage,
  unknown arguments will be silently ignored. Of course you can also use
  an ArgChecker that accepts unknown options to serve as a catch-all.
  This ArgChecker evaluates option.Value which Parse() uses to communicate
  additional information in the following manner:
  
    - if option.Value is a 2-element []string, then the option is an
      ambiguous abbreviation of (at least) the 2 option names in the []string
      ArgUnknown will report the 2 names in the returned error message.
    
*/
func ArgUnknown(option *Option) error {
  if ambiguous, ok := option.Value.([]string); ok && len(ambiguous) == 2 {
    return fmt.Errorf("Ambiguous abbreviation '%v'. Candiates: --%v, --%v", option, ambiguous[0], ambiguous[1])
  }
  return fmt.Errorf("Unknown option '%v'", option)
}

/*
  ArgChecker that always fails with an error that says the option is unimplemented.
  Useful as placeholder in Usage.
*/
func ArgUnimpl(option *Option) error {
  return fmt.Errorf("Option '%v' not implemented", option)
}

/*
  ArgChecker that accepts any argument as long as there is one.
  Note that the empty string is a valid argument for this ArgChecker
  (e.g. --title= ).
*/
func ArgRequired(option *Option) error {
  if option.HasArg {
    return ARG_OK
  }
  return fmt.Errorf("Option '%v' requires an argument", option)
}

/*
  ArgChecker that accepts any argument as long as it is not the
  empty string. E.g. --file="" would be rejected by this ArgChecker.
*/
func ArgNonEmpty(option *Option) error {
  if option.HasArg && option.Arg != "" {
    return ARG_OK
  }
  return fmt.Errorf("Option '%v' requires a non-empty argument", option)
}

/*
  ArgChecker that accepts base-10 integers and stores them as int in option.Value.
*/
func ArgInt(option *Option) error {
  if option.HasArg {
    i, err := strconv.Atoi(option.Arg)
    if err == nil {
      option.Value = i
      return ARG_OK
    }
  }
  return fmt.Errorf("Option '%v' requires an integer as argument", option)
}

/*
  ArgChecker for options that do not have an argument.
*/
func ArgNone(option *Option) error {
  return ARG_NONE_Err(option.Name)
}

/*
  ArgChecker for options that may or may not have an argument.
  This ArgChecker accepts only attached arguments, because of the
  ambiguities that would otherwise arise (e.g. What does "-f -g" mean
  if -f takes an optional argument?)
*/
func ArgOptional(option *Option) error {
  if option.HasArg {
    if option.ArgAttached {
      return ARG_OK
    } else {
      return ARG_NONE{fmt.Errorf("Option %v only accepts attached arguments", option.Name)}
    }
  }
  return ARG_NONE{fmt.Errorf("No argument given (but that's ok, it's optional)")}
}

/*
  Takes an argument vector (typically os.Args[1:]) and parses it according to
  usage and flags.
  
   flags contains zero or more words separated by space. The following words
   are understood:
   
    "gnu": By default Parse() uses the behavior specified by POSIX for getopt()
           which is to treat every word in the argument vector after the
           first non-option word as a non-option even if it starts with "-".
           The "gnu" flag changes Parse() behavior to that of GNU getopt() which
           continues parsing for options even after a non-option until the
           special "--" separator is encountered (after which definitely every
           word is a non-option no matter what it looks like).
         
    "-perl": This makes Parse() accept Perl-style single-minus long options.
             The double minus form will still be recognized. Note that
             single minus long options take precedence over short options and
             short option groups. E.g. "-file" would be interpreted as "--file"
             and not as "-f -i -l -e" (assuming a long option named "file" exists).
                      
    "--a", "--ab", "--abb", "--abbr", ... "--abbreviated":
           Any word starting with "--a" will tell Parse() to recognize abbreviated
           long options as long as the abbreviated form has at least as many characters
           as the "-a..." word. Parse() will match a prefix of a long option as if it was
           the full long option (e.g. "--foob=10" will be interpreted as if it was
           "--foobar=10" ), as long as the prefix has the required length AND IS UNAMBIGUOUS.
           
    Be careful if combining "-perl" and "--a" (i.e. abbreviations with only a 1 character
    prefix) because the ambiguity check does not consider short options and abbreviated
    single minus long options will take precedence over short options.
                                 
  The returned values are
  
    options: a slice with one entry for each non-negative OptionInfo.Id in usage.
             options[id] is nil if none of the arguments described in usage with that id
             is found in the argument vector. Otherwise it's a pointer to the parsed
             Option corresponding to the first option in the argument vector with that id.
             From that entry you can follow the Next() chain to iterate over all
             options with the same OptionInfo.Id.
    
    nonoptions: a list of all non-option words in the argument vector (excluding the "--"
                separator if present). Note that in POSIX mode (the default, see flags)
                the first argument that doesn't start with '-' or '--' and does not belong to
                a preceding argument-taking option, will terminate the option list and is the
                first non-option argument. All following words in the argument vector
                are treated as non-option arguments, even if they start with '-' .
    
    err: if non-nil something went wrong and the other return values have unspecified values.
    
    alloptions: contains all the parsed options from the argument vector in order, i.e.
                alloptions[i] corresponds to the i-th option in the argument vector.
                Note the difference to the returned options slice which is indexed
                by the OptionInfo.Id values. The alloptions slice is used to iterate
                over all options found in the argument vector. The options slice is used
                to access specific options directly.
                NOTE: Each option from the argument vector produces only one Option
                structure. The same Option structures are referenced in the
                returned options and alloptions slices.
*/
func Parse(args []string, usage Usage, flags string) (options []*Option, nonoptions []string, err error, alloptions []*Option) {
  maxindex := 0
  for _, info := range usage {
    if info.Id > maxindex { maxindex = info.Id }
  }
  options = make([]*Option, maxindex+1)
  
  gnu := false
  single_minus_longopt := false
  min_abbr_len := 0
  
  for _, flg := range strings.Fields(flags) {
    if flg == "gnu" {
      gnu = true
    } else if flg == "-perl" {
      single_minus_longopt = true
    } else if strings.HasPrefix(flg, "--abb") {
      min_abbr_len = len(flg) - 2 // -2 for the "--"
    } else {
      err = fmt.Errorf("Parse(): Unknown word in flags: %v", flg)
      return
    }
  }
  
  numargs := len(args)

  argidx := 0

  for numargs > 0 {
    param := args[argidx] // param can be --long-option, -srto or non-option argument

    // in POSIX mode the first non-option argument terminates the option list
    // a lone minus character is a non-option argument
    if param == "" || param == "-" || param[0] != '-' {
      if gnu {
        nonoptions = append(nonoptions, param)
        argidx++
        numargs--
        continue
      } else {
        break
      }
    }

    // -- terminates the option list. The -- itself is ignored and not appended to nonoptions
    if param == "--" {
      argidx++
      numargs--
      break
    }
    
    // param is not "", "-" or "--" at this point and we know param[0] == '-'.
    // This guarantees that len(param) >= 2 if it starts with a single '-' 
    // and >= 3 if it starts with "--"

    var handle_short_options bool
    var longopt_name string
    if param[1] == '-' { // if --long-option
      handle_short_options = false
      longopt_name = param[2:]
    } else {
      handle_short_options = true
      longopt_name = param[1:] //for testing a potential -long-option
    }

    try_single_minus_longopt := single_minus_longopt
    have_more_args := numargs > 1

    // loop over short options in group, for long options the body is executed only once
    for {
      var idx int

      var optarg string
      have_optarg := true
      attached_arg := false
      separate_arg := false
      ambiguous := []string{}
      bettername := ""
      use_bettername := try_single_minus_longopt
      
      /******************** long option **********************/
      if (handle_short_options == false || try_single_minus_longopt) {
        idx = 0
        for idx < len(usage) && !streq(usage[idx].Long, longopt_name) {
          idx++
        }

        if idx >= len(usage) && min_abbr_len > 0 { // if we should try to match abbreviated long options
          i1 := 0
          for i1 < len(usage) && !streqabbr(usage[i1].Long, longopt_name, min_abbr_len) {
            i1++
          }
          if i1 < len(usage) { // now test if the match is unambiguous by checking for another match
            ambiguous = append(ambiguous, usage[i1].Long)
            i2 := i1 + 1
            for i2 < len(usage) && !streqabbr(usage[i2].Long, longopt_name, min_abbr_len) {
              i2++
            }

            if i2 >= len(usage) { // if there was no second match it's unambiguous, so accept i1 as idx
              idx = i1
            } else {
              ambiguous = append(ambiguous, usage[i2].Long)
            }
          }
        }

        try_single_minus_longopt = false // prevent looking for longopt in the middle of shortopt group

        // if we found something
        if idx < len(usage) {
          // disable handle_short_options (only relevant if single_minus_longopt)
          handle_short_options = false

          for optarg = longopt_name; optarg != "" && optarg[0] != '='; {
            optarg = optarg[1:]
          }
          if optarg != "" && optarg[0] == '=' { // attached argument
            optarg = optarg[1:] // chop off "="
            attached_arg = true
          } else {
            // possibly detached argument
            if have_more_args {
              optarg = args[argidx+1]
              separate_arg = true
            } else {
              have_optarg = false
            }
          }
        }
      }

      /************************ short option ***********************************/
      if handle_short_options {
        param = param[1:] // point at the 1st/next option character
        if param == "" {
          break // end of short option group
        }

        idx = 0
        for idx < len(usage) && !instr(param[0], usage[idx].Short) {
          idx++
        }

        if len(param) == 1 { // if the potential argument is separate
          if have_more_args {
            optarg = args[argidx+1]
            separate_arg = true
          } else {
            have_optarg = false
          }
        } else {
          // if the potential argument is attached
          optarg = param[1:]
          attached_arg = true
        }
      }

      var descriptor *OptionInfo
      
      if idx >= len(usage) { /**************  unknown option ********************/
        // look for dummy entry (Short == "" and Long == "") to use as descriptor for unknown options
        idx = 0
        for idx < len(usage) && (usage[idx].Short != "" || usage[idx].Long != "") {
          idx++
        }
        if use_bettername && (param[0] != '-') {
          bettername = "-" + longopt_name
        }
      }
      
      if idx < len(usage) {
        descriptor = &usage[idx]
        
        var name string
        if param[0] != '-' { // short option
          name = "-" + param[0:1]
        } else {
          if eqidx := strings.Index(param, "="); eqidx >= 0 {
            name = param[0:eqidx]
          } else {
            name = param
          }
        }

        option := &Option{
          Name: name,
          HasArg: have_optarg,
          Arg: optarg,
          ArgAttached: attached_arg,
          Info: descriptor,
        }
        
        if bettername != "" {
          option.Name = bettername
        }
        
        if len(ambiguous) == 2 {
          option.Value = ambiguous
        }
        
        err = descriptor.CheckArg(option)
        
        if err == ARG_OK {
          // skip one element of the argument vector, if it's a separated argument
          if have_optarg && have_more_args && separate_arg {
            numargs--
            argidx++
          }

          // No further short options are possible after an argument
          handle_short_options = false;
          
        } else if _, ok := err.(ARG_NONE); ok {
          if attached_arg && !handle_short_options { // if the argument is attached to a long option, we can't just ignore it
            return          // => pass error to caller
          }
          option.HasArg = false
          option.Arg = ""
        } else {
          return
        }
        
        alloptions = append(alloptions, option)
        id := option.Info.Id
        if id >= 0 {
          if options[id] == nil {
            options[id] = option
            option.next = option
          } else {
            last := options[id].Last()
            last.isNotLast = true
            option.next = last.next
            last.next = option
          }
        }
      }

      if !handle_short_options { break }
    }

    argidx++
    numargs--
  }

  for numargs > 0 {
    nonoptions = append(nonoptions, args[argidx])
    argidx++
    numargs--
  }

  err = nil
  return
}


func (usage Usage) String() string {
  return formatUsage(usage)
}

/*
 * Returns true iff character ch is contained in the string st.
 */
func instr(ch byte, st string) bool {
  for i := 0; i < len(st); i++ {
    if st[i] == ch { return true }
  }
  return false
}

/*
 * Returns true iff st1 is a prefix of st2 and
 * in case st2 is longer than st1, then
 * the first additional character is '='.
 *
 * Examples:
 *
 *   streq("foo", "foo=bar") == true
 *   streq("foo", "foobar")  == false
 *   streq("foo", "foo")     == true
 *   streq("foo=bar", "foo") == false
 */
func streq(st1, st2 string) bool {
  if len(st1) > len(st2) {
    return false
  }
  if len(st1) < len(st2) && st2[len(st1)] != '=' {
    return false
  }
  if st1 != st2[0:len(st1)] {
    return false
  }
  return true
}

/**
 * Like streq() but handles abbreviations.
 *
 * Returns true iff st1 and st2 have a common
 * prefix with the following properties:
 * 
 *  - (if min > 0) its length is at least min characters
 *    or the same length as st1 (whichever is smaller).
 *  - (if min <= 0) its length is the same as that of st1
 *  - within st2 the character following the common prefix
 *    is '=' or st2 is identical to the common prefix.
 *
 * Examples:
 *
 *   streqabbr("foo", "foo=bar",<anything>) == true
 *   streqabbr("foo", "fo=bar" , 2) == true
 *   streqabbr("foo", "fo"     , 2) == true
 *   streqabbr("foo", "fo"     , 0) == false
 *   streqabbr("foo", "f=bar"  , 2) == false
 *   streqabbr("foo", "f"      , 2) == false
 *   streqabbr("fo" , "foo=bar",<anything>)  == false
 *   streqabbr("foo", "foobar" ,<anything>)  == false
 *   streqabbr("foo", "fobar"  ,<anything>)  == false
 *   streqabbr("foo", "foo"    ,<anything>)  == true
 */
func streqabbr(st1, st2 string, min int) bool {
  i1 := 0
  i2 := 0
  for i1 != len(st1) && i2 != len(st2) && st1[i1] == st2[i2] {
    i1++
    i2++
  }

  return (i1 == len(st1) || (min > 0 && i1 >= min)) && (i2 == len(st2) || st2[i2] == '=')
}
