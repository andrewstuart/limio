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
DefaultWindow is the window used to smooth SimpleLimit rates. That is,
SimpleLimit distributes the given quantity evenly into buckets of size t. This
is useful for avoiding tcp silly window syndrome and providing predictable
resource usage.

```go
var ErrTimeoutExceeded error = errors.New("Timeout Exceeded")
```
ErrTimeoutExceeded will be returned upon a timeout lapsing without a read
occuring

#### func  Distribute

```go
func Distribute(n int, t, w time.Duration) (int, time.Duration)
```
Distribute takes a rate (n, t) and window (w), evenly distributes the n/t to
n'/t' (n'<=n && t'>=w)

#### type Limiter

```go
type Limiter interface {
	Limit(chan int) <-chan bool //The channel is useful for knowing that the channel has been unlimited. The boolean represents finality.
	Unlimit()
}
```

A Limiter is an interface that meters some underlying discretely quantifiable
operation with respect to time.

The Limit() function, when implemented, should apply a limit to some underlying
operation when called. Supporting concurrency is up to the implementer and as
such, should be documented. The semantics of the channel are that of a token
bucket. The actual integer sent through the channel represents a quantity of
operations that can take place. The implementation should be sure to specify its
interpretation of the quantity.

Unlimit() removes any formerly imposed limits and allows the underlying
operation

#### type Manager

```go
type Manager interface {
	Limiter
	Manage(Limiter) error
	Unmanage(Limiter)
}
```

A Manager enables consumers to treat a group of Limiters as a single Limiter,
enabling hierarchies of limiters. For example, a network interface could have a
global limit that is distributed across connections, each of which can manage
their own distribution of the bandwidth they are allocated.

#### type Reader

```go
type Reader struct {
}
```

Reader implements an io-limited reader that conforms to the io.Reader and
limio.Limiter interface, and can have its limits updated concurrently with any
Read() calls.

#### func  NewReader

```go
func NewReader(r io.Reader) *Reader
```
NewReader takes any io.Reader and returns a limio.Reader.

#### func (*Reader) Close

```go
func (r *Reader) Close() error
```
Close allows the goroutines that were managing limits and reads to shut down and
free up memory. It should be called by any clients of the limio.Reader, much as
http.Response.Body should be closed to free up system resources.

#### func (*Reader) Limit

```go
func (r *Reader) Limit(lch chan int) <-chan bool
```
Limit can be used to precisely control the limit at which bytes can be Read,
whether burstily or not.

#### func (*Reader) Read

```go
func (r *Reader) Read(p []byte) (written int, err error)
```
Read implements io.Reader in a blocking manner according to the limits of the
limio.Reader.

#### func (*Reader) SetTimeout

```go
func (r *Reader) SetTimeout(t time.Duration) error
```
SetTimeout takes some time.Duration t and configures the underlying Reader to
return a limio.TimedOut error if the timeout is exceeded while waiting for a
read operation.

#### func (*Reader) SimpleLimit

```go
func (r *Reader) SimpleLimit(n int, t time.Duration) <-chan bool
```
SimpleLimit takes an integer and a time.Duration and limits the underlying
reader non-burstily (given rate is averaged over a small time).

#### func (*Reader) Unlimit

```go
func (r *Reader) Unlimit()
```
Unlimit removes any restrictions on the underlying io.Reader.

#### type SimpleManager

```go
type SimpleManager struct {
}
```

A SimpleManager is an implementation of the limio.Manager interface. It allows
simple rate-based and arbitrary channel-based limits to be set.

#### func  NewSimpleManager

```go
func NewSimpleManager() *SimpleManager
```
NewSimpleManager creates and initializes a SimpleManager.

#### func (*SimpleManager) Close

```go
func (lm *SimpleManager) Close() error
```
Close allows the SimpleManager to free any resources it is using if the consumer
has no further need for the SimpleManager.

#### func (*SimpleManager) Limit

```go
func (lm *SimpleManager) Limit(l chan int) <-chan bool
```
Limit implements the limio.Limiter interface.

#### func (*SimpleManager) Manage

```go
func (lm *SimpleManager) Manage(l Limiter) error
```
Manage takes a Limiter that will be adopted under the management policy of the
SimpleManager.

#### func (*SimpleManager) NewReader

```go
func (lm *SimpleManager) NewReader(r io.Reader) *Reader
```
NewReader takes an io.Reader and Limits it according to its limit
policy/strategy

#### func (*SimpleManager) SimpleLimit

```go
func (lm *SimpleManager) SimpleLimit(n int, t time.Duration) <-chan bool
```
SimpleLimit takes an int and time.Duration that will be distributed evenly
across all managed Limiters.

#### func (*SimpleManager) Unlimit

```go
func (lm *SimpleManager) Unlimit()
```
Unlimit implements the limio.Limiter interface.

#### func (*SimpleManager) Unmanage

```go
func (lm *SimpleManager) Unmanage(l Limiter)
```
Unmanage allows consumers to remove a specific Limiter from its management
strategy
