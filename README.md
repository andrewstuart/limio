[![GoDoc](https://godoc.org/github.com/andrewstuart/limio?status.svg)](https://godoc.org/github.com/andrewstuart/limio)

# limio
--
    import "git.astuart.co/andrew/limio"

Package limio provides an interface abstraction for rate limiting or flow
control of arbitrary io.Readers or io.Writers. A concrete implementation of a
handler is also provided for ease of use and reference are also provided.

## Usage

```go
const (
	B int = 1 << (10 * (iota))
	KB
	MB
	GB
	TB
	PB
	EB
)
```
Some useful byte-sized (heh) constants

```go
const DefaultWindow = 10 * time.Millisecond
```

#### func  Distribute

```go
func Distribute(n int, t, w time.Duration) (int, time.Duration)
```
Distribute takes a rate (n, t) and window (w), evenly distributes the n/t to
n'/t' (n'<=n && t'>=w)

#### type Limiter

```go
type Limiter interface {
	Limit(n int, t time.Duration) <-chan bool //The channel is useful for knowing that the channel has been unlimited
	LimitChan(chan int) <-chan bool
	Unlimit()
	io.Closer
}
```


#### type Manager

```go
type Manager interface {
	Limiter
	Manage(Limiter)
	Unmanage(Limiter)
}
```


#### type Reader

```go
type Reader struct {
}
```


#### func  NewReader

```go
func NewReader(r io.Reader) *Reader
```

#### func (*Reader) Close

```go
func (r *Reader) Close() error
```

#### func (*Reader) Limit

```go
func (r *Reader) Limit(n int, t time.Duration) <-chan bool
```

#### func (*Reader) LimitChan

```go
func (r *Reader) LimitChan(lch chan int) <-chan bool
```

#### func (*Reader) Read

```go
func (r *Reader) Read(p []byte) (written int, err error)
```

#### func (*Reader) Unlimit

```go
func (r *Reader) Unlimit()
```

#### type SimpleManager

```go
type SimpleManager struct {
}
```


#### func  NewSimpleManager

```go
func NewSimpleManager() *SimpleManager
```

#### func (*SimpleManager) Close

```go
func (lm *SimpleManager) Close() error
```

#### func (*SimpleManager) Limit

```go
func (lm *SimpleManager) Limit(n int, t time.Duration) <-chan bool
```

#### func (*SimpleManager) LimitChan

```go
func (lm *SimpleManager) LimitChan(l chan int) <-chan bool
```

#### func (*SimpleManager) Manage

```go
func (lm *SimpleManager) Manage(l Limiter)
```

#### func (*SimpleManager) NewReader

```go
func (lm *SimpleManager) NewReader(r io.Reader) *Reader
```

#### func (*SimpleManager) Unlimit

```go
func (lm *SimpleManager) Unlimit()
```

#### func (*SimpleManager) Unmanage

```go
func (lm *SimpleManager) Unmanage(l Limiter)
```
