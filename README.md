# golib
Library of utility code for Go programs

## util
Various utility functions. Of particular interest are the logging and base64 functions.

## deque
A data structure that can serve as FIFO, LIFO, vector, array, list, stack, queue, buffer,
channel, inter-process-communication mechanism. It's synchronized so that it can be used by
multiple goroutines in parallel. It can be used blocking (with timeouts) and non-blocking.
Its speed and memory-usage are appropriate for most situations.
Basically, whenever you need a non-tree data structure, just use this one, and only consider
something else after you've tried it and have encountered a specific reasons why it doesn't work
in that situation.

## bytes
An alternative to bytes/buffer from the standard Go library. Unlike the standard version this one
is based on malloc()/free(). This is useful when you are dealing with very large buffers and
need to control their lifetime manually to avoid out-of-memory issues.
