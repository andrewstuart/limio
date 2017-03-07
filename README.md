[![GoDoc](https://godoc.org/github.com/andrewstuart/limio?status.svg)](https://godoc.org/github.com/andrewstuart/limio)
[![Build Status](https://travis-ci.org/andrewstuart/limio.svg?branch=master)](https://travis-ci.org/andrewstuart/limio)

Limio is meant to be a dead simple rate-limiting library, primarily aimed at
providing an intuitive API surface for composable operational constraints. By
centering around `chan int`, limio provides token bucket implementations with
decent backpressure and helps eliminate silly window syndrome by spreading out
token distribution over smaller units of time to provide a more even flow.

By using a channel, we also eliminate the need of Limiter implementations to
follow a complex contract involving rates. They simply need to listen on a 
channel for some quantity, and then when a quantity arrives, perform that many
operations.

Some examples usage of the limio.Reader

```go
func slowCopy(w io.Writer, r io.Reader) error {
  lr := limio.NewReader(r)

  // Limit at 1MB/s
  lr.SimpleLimit(1*limio.KB, time.Second)

  return io.Copy(w, lr)
}
```
