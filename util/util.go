/* Copyright (C) 2013 Matthias S. Benkmann
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this file (originally named util.go) and associated documentation files 
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
package util // import "winterdrache.de/golib/util"

import (
         "io"
         "fmt"
         "net"
         "time"
         "regexp"
         "strings"
         "crypto/md5"
         "runtime/debug"
       )

// Returns the md5sum of its argument as a string of hex digits.
func Md5sum(s string) string {
  md5 := md5.New()
  io.WriteString(md5, s)
  return fmt.Sprintf("%x",md5.Sum(nil))
}

// Number of consecutive short writes before WriteAll() will give up
// ATTENTION! The wait time between tries increases exponetially, so don't
// blindly increase this number.
const write_all_max_tries = 8

// Calls g wrapped in a panic handler that logs the panic and recovers from it.
// Example:
//   go util.WithPanicHandler(foobar)
//   go util.WithPanicHandler(func(){ Send_foreign_job_updates(server, jobs) })
func WithPanicHandler(g func()) {
  defer func() {
    if x := recover(); x != nil {
      Log(0, "PANIC! %v\n%v", x, string(debug.Stack()))
    }
  }()
  g()
}

// Returns a new channel that will return start as first item and then
// always the next number.
func Counter(start uint64) chan uint64 {
  c := make(chan uint64)
  go func() {
    for {
      c<-start
      start++
    }
  }()
  return c
}

// Sleeps until time t. If clock is adjusted during the sleep, this will not
// cause more than 10 minutes extra sleep. In no case will this function return
// until time.Now() >= t.
func WaitUntil(t time.Time) {
  // Wait in a for loop because clock might be adjusted while we're sleeping.
  // ATTENTION! The check t.After(time.Now()) is NOT REDUNDANT with the check
  // if dur <= 0 { return }
  // A time.Duration has a limited range that can not express all differences
  // between timestamps. Without the t.After(time.Now()) check, extreme timestamps
  // in the past can cause large wait times due to overflow.
  for ; t.After(time.Now()); { 
    dur := t.Sub(time.Now())
    if dur <= 0 { return }
    // wake up every 10 minutes to deal with clock adjustments (DST etc.)
    if dur > 10*time.Minute { dur = 10*time.Minute }
    time.Sleep(dur)
  }
}

// Waits until either the duration timeout has passed or DNS is available
// ON A NON-LOOPBACK INTERFACE (so this waits until external interfaces are up).
// If timeout == 0, wait forever if necessary.
// Returns true if DNS is available.
func WaitForDNS(timeout time.Duration) bool {
  var err error
  
  endtime := time.Now().Add(timeout)
  for {
    var ifaces []net.Interface
    ifaces, err = net.Interfaces()
    if err == nil {
      // find non-loopback interfaces that are up
      for _, iface := range ifaces {
        if iface.Flags & net.FlagLoopback != 0 { continue }
        if iface.Flags & net.FlagUp == 0 { continue }
        
        var addrs []net.Addr
        addrs, err = iface.Addrs()
        if err == nil {
          
          // try to find a non-loopback IP address for that interface with rDNS
          for _, addr := range addrs {
            ip, _, err2 := net.ParseCIDR(addr.String())
            if err2 == nil && !ip.IsLoopback() {
              var names []string
              names, err = net.LookupAddr(ip.String())
              if err == nil && len(names) > 0 {
                // sanity-check: lookup an address in the special reserved TEST IP range
                //               (this should never resolve to a name)
                //               and lookup the loopback address 127.0.0.1 (should always resolve)
                _, err192 := net.LookupAddr("192.0.2.1")
                _, err127 := net.LookupAddr("127.0.0.1")
                if err192 != nil && err127 == nil { return true }
              }
            }
          }
        }
      }
    }
    
    if timeout != 0 && time.Now().After(endtime) { break }
    waittime := endtime.Sub(time.Now())
    if waittime <= 0 || waittime > 1*time.Second { waittime = 1*time.Second }
    time.Sleep(waittime)
  }
  return false
}

// Writes data to w, with automatic handling of short writes.
// A short write error will only be returned if multiple attempts
// failed in a row.
func WriteAll(w io.Writer, data []byte) (n int, err error) {
  // Yeah, I know. Other people just ignore the issue of
  // short writes. That's why their code fails more often than mine :-P
  tries := write_all_max_tries
  var bytes_written int
  for n = 0; n < len(data); {
    bytes_written, err = w.Write(data[n:])
    n += bytes_written
    
    if err != nil && err != io.ErrShortWrite {
      return n, err
    }
    
    if bytes_written == 0 {
      tries--
      if tries <= 0 {
        if err == nil {
          err = io.ErrShortWrite
        }
        return n, err
      }
      
      // The first time we don't sleep. The 2nd time we sleep 1ms. The 3rd time 2ms.
      // The 4th time 4ms. Then 8ms, 16ms, 32ms, 64ms,...
      var wait time.Duration = (1 << (write_all_max_tries-2)) >> uint(tries)
      time.Sleep(wait * time.Millisecond)
      
    } else {
      tries = write_all_max_tries  // every time we succeed at writing we start tries again
    }
  }
  
  return n, nil
}

// Opens a connection to target (e.g. "foo.example.com:20081"), sends msg followed
// by \r\n and then closes the connection.
// If timeout >= 0, then the connection will be terminated after at most this duration.
func SendLnTo(target string, msg string, timeout time.Duration) {
  conn, err := net.Dial("tcp", target)
  if err != nil {
    Log(0, "ERROR! Dial: %v", err)
    return
  }
  defer conn.Close()
  SendLn(conn, msg, timeout)
}

// Sends strings via connection conn, followed by "\r\n"
// If timeout >= 0, then the send attempt will be aborted after at most this duration.
// Returns nil if sending was successful, an error otherwise. The error will also
// be sent to the log, so the caller does not have to log it.
func SendLn(conn net.Conn, s string, timeout time.Duration) error {
  sendbuf := make([]byte, len(s)+2)
  copy(sendbuf, s)
  sendbuf[len(s)]='\r'
  sendbuf[len(s)+1]='\n'

  var deadline time.Time // zero value means "no deadline"
  if timeout >= 0 { deadline = time.Now().Add(timeout) }
  conn.SetWriteDeadline(deadline)
  var no_deadline time.Time
  defer conn.SetWriteDeadline(no_deadline)
  
  _, err := WriteAll(conn, sendbuf)
  if err != nil {
    Log(0, "ERROR! WriteAll: %v", err)
  }
  return err
}

// Reads from the connection until \n is seen (or timeout or error) and
// returns the read data without any trailing \n or \r.
//
// If timeout > 0, then reading will stop after at most this duration and
// the function will return with an error. If timeout <= 0 there will be
// no timeout (any pre-existing read deadline will be cleared).
//
// If an error (or timeout) occurs before \n is seen, the data read up to
// the error is returned (without any trailing \r).
//
// You should probably give special treatment to the case of an io.EOF error
// being returned. This happens when the last line of the connection is not
// terminated by \n. It often makes sense to not treat this as an error.
//
// NOTE: This function reads individual bytes which means that data following
// the line read is still available for reading (in particular you can read
// subsequent lines by calling ReadLn() again). The drawback, however, is
// lower throughput compared to reading in chunks.
func ReadLn(conn net.Conn, timeout time.Duration) (string, error) {
  var deadline time.Time // zero value means "no deadline"
  if timeout > 0 { deadline = time.Now().Add(timeout) }
  conn.SetReadDeadline(deadline)  
  var no_deadline time.Time
  defer conn.SetReadDeadline(no_deadline)
  
  var buf = make([]byte, 128)
  var i int
  var n int
  var err error
  
  for err == nil {
    if i == len(buf) {
      buf_new := make([]byte, len(buf)<<1)
      copy(buf_new, buf)
      buf = buf_new
    }
    
    n, err = conn.Read(buf[i:i+1]) // read one byte
    
    i += n

    // Did we read \n ?
    if i > 0 && buf[i-1] == '\n' {
      for ; i > 0 && (buf[i-1] == '\n' || buf[i-1] == '\r') ; { i-- }
      return string(buf[0:i]), nil
    }
  }
  
  for ; i > 0 && buf[i-1] == '\r' ; { i-- }
  return string(buf[0:i]), err
}

// Sends a UDP wake-on-lan packet coded for the given macaddress to the
// broadcast address for the given host or network.
//  host_or_net can be:
//    a) a host name (with or without ":<port>")
//    b) a host's IP address from which the broadcast address will be
//       derived by setting final byte to 255
//    c) a network's broadcast IP address 
//  If host_or_net includes a port that port will be used, otherwise the
//  default port 40000 is used.
func Wake(macaddress string, host_or_net string) error {
  hwaddr, err := net.ParseMAC(macaddress)
  if err != nil { return err }

  if strings.Index(host_or_net,":") < 0 { host_or_net += ":40000" }
  
  udpaddr,err := net.ResolveUDPAddr("udp", host_or_net)
  if err != nil { return err }
  
  udpaddr.IP[len(udpaddr.IP)-1] = 255
  udpconn,err := net.DialUDP("udp", nil, udpaddr)
  if err != nil { return err }
    
  payload := []byte{0xff,0xff,0xff,0xff,0xff,0xff}
  for i := 0; i < 16; i++ { payload = append(payload, hwaddr...) }
  _, err = udpconn.Write(payload)
  return err
}

const re_1xx = "(1([0-9]?[0-9]?))"
const re_2xx = "(2([6-9]|([0-4][0-9]?)|(5[0-5]?))?)"
const re_xx  = "([3-9][0-9]?)"
const re_port = "(0|([1-6][0-9]{0,4})|([7-9][0-9]{0,3}))"
const ip_part = "(0|"+re_1xx+"|"+re_2xx+"|"+re_xx+")"
var startsWithIPv4Regexp = regexp.MustCompile("^"+ip_part+"([.]"+ip_part+"){3}")
var endsWithPort = regexp.MustCompile(":"+re_port+"$")

// Takes either a HOST or a HOST:IP pair and replaces the HOST part with
// that machine's IP address (preferring an IPv4 address if there is one).
// If HOST is "localhost" or "::1", it will be replaced with "127.0.0.1".
// If HOST is already an IPv4 address, it will be kept.
// If there is only an IPv6 address available, it will be enclosed in "[...]"
// in the result (even if there is no port).
//
// If localip != "" in the result "127.0.0.1" will be replaced with localip.
// 
// Returns the modified address or the original address with an error.
func Resolve(addr string, localip string) (string, error) {
  if localip == "" { localip = "127.0.0.1" }
  addr = strings.Replace(addr, "127.0.0.1", localip, 1)
  
  if startsWithIPv4Regexp.MatchString(addr) { return addr, nil }
  
  host := addr
  port := ""
  var err error

  // the net.ParseIP() check tries to prevent confusing an IPv6 address for a port
  if endsWithPort.MatchString(addr) && net.ParseIP(addr) == nil {
    host, port, err = net.SplitHostPort(addr)
    if err != nil { return addr, err }
    port = ":" + port
  }
  
  if host == "localhost" || host == "::1" || host == "[::1]" { 
    return localip + port, nil 
  }
  
  addrs, err := net.LookupIP(host)
  if err != nil {
    return addr, err
  }
  
  if len(addrs) == 0 { // I don't think this is possible but just in case...
    return addr, fmt.Errorf("No IP address for %v", host)
  }
  
  // try to find an IPv4 non-loopback address
  for _, a := range addrs {
    if !a.IsLoopback() && a.To4() != nil { return a.String() + port, nil }
  }
  
  // try to find an IPv4 address (possibly loopback)
  for _, a := range addrs {
    if a.To4() != nil { return strings.Replace(a.String(), "127.0.0.1", localip, 1) + port, nil }
  }
  
  // take the first address (which is IPv6)
  return "[" + addrs[0].String() + "]" + port, nil
}

