/* Copyright (C) 2013 Matthias S. Benkmann
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this file (originally named timestamp.go) and associated documentation files 
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

import "time"
import "fmt"
import "strings"
import "strconv"

// Converts t into a timestamp appropriate for use in siserver messages.
// The timestamp loses the time zone information of t. No time zone
// conversion will be performed. IOW "12:00 UTC" and "12:00 PDT" will
// both result in a timestamp that says "yyyymmdd1200ss".
func MakeTimestamp(t time.Time) string {
  return t.Format("20060102150405")
}

// Converts a timestamp as used in siserver messages into a time.Time.
// The returned time will be the time at which the server clock's current
// time converted with MakeTimestamp() is ts. The computation is based on
// the assumption that the server's time zone does not change EXCEPT for
// daylight savings time. IOW on a server running on local time in Berlin
// ParseTimestamp("20140101120000") gives 12:00 CET (winter time) and
// ParseTimestamp("20140601120000") gives 12:00 CEST (summer time).
//
// ParseTimestamp() returns time.Unix(0,0) if the timestamp is invalid.
func ParseTimestamp(ts string) time.Time {
  t, err := time.Parse("20060102150405", ts)
  if err != nil {
    Log(0, "ERROR! Illegal timestamp: %v (%v)", ts, err)
    return time.Unix(0,0)
  }
  
  // The timestamp doesn't contain information about the time zone it
  // is in, so time.Parse() has interpreted it as UTC. In order to
  // properly get a time in local time, we need to first determine
  // the code of the time zone active at the requested time. Then
  // we reparse with that time zone code appended.
  zone,_ := t.Local().Zone()
  ts += zone
  t, err = time.Parse("20060102150405MST", ts)
  if err != nil {
    Log(0, "ERROR! Could not parse with time zone: %v (%v)", ts, err)
    return time.Unix(0,0)
  }
  return t
}

// Takes a timestamp ts in the format produced by MakeTimestamp and
// adds to it an offset adder in the following
// format:
//   <integer>_seconds
//   <integer>_minutes
//   <integer>_hours
//   <integer>_days
//   <integer>_weeks
//   <integer>_months
//   <integer>_years
//
// where <integer> is any integer (may be negative).
// Example: AddTimestamp("20170201120000", "-3_seconds")
//          returns "20170201115957"
//
// If ts is invalid, time.Unix(0,0) is assumed and an error
// is logged (behavior of ParseTimestamp()).
// If adder is invalid, the timestamp is returned unchanged
// together with an error.
func AddTimestamp(ts string, adder string) (string, error) {
  t := ParseTimestamp(ts)
  p := strings.Split(adder, "_")
  if len(p) != 2 {
    return ts, fmt.Errorf("Invalid timestamp offset: \"%v\"", adder)
  }
  offset, err := strconv.ParseInt(p[0], 10, 64)
  if err != nil {
    return ts, fmt.Errorf("Invalid timestamp offset: \"%v\": %v", adder, err)
  }
  
  switch p[1] {
    case "seconds": t = t.Add(time.Duration(offset) * time.Second)
    case "minutes": t = t.Add(time.Duration(offset) * time.Minute)
    case "hours":   t = t.Add(time.Duration(offset) * time.Hour)
    case "days":    t = t.AddDate(0,0,int(offset))
    case "weeks":   t = t.AddDate(0,0,int(offset*7))
    case "months":  t = t.AddDate(0,int(offset),0)
    case "years":   t = t.AddDate(int(offset),0,0)
    default:
         return ts, fmt.Errorf("Unknown timestamp offset unit: \"%v\"", p[1])
  }
  
  return MakeTimestamp(t), nil
}
