# ratelimit

[Docs available on godoc](https://godoc.org/github.com/andrewstuart/ratelimit)
PACKAGE DOCUMENTATION

package ratelimit
    import "."

    Package ratelimit provides an interface abstraction for rate limiting or
    flow control of arbitrary io.Readers or io.Writers. Several concrete
    implementations of Limiters are also provided.

FUNCTIONS

func NewReader(r io.Reader, l Limiter) io.Reader
    NewReader takes an io.Reader and a Limiter and returns an io.Reader that
    will be limited via the strategy that Limiter choses to implement.

TYPES

type ByteCount uint64
    A ByteCount is simply an abstraction over some integer type to provide
    more flexibility should the type need to be changed. Since zettabyte
    overflows uint64, I suppose it may someday need to be changed.

const (
    B ByteCount = 1 << (10 * (iota))
    KB
    MB
    GB
    TB
    PB
    EB
)
    Some useful constants with the proper typing

type Limiter interface {
    GetLimit() <-chan ByteCount
}
    A Limiter should implement some strategy for providing access to a
    shared io resource. The GetLimit() function must return a channel of
    ByteCount. When it is appropriate for the new limited io.Reader to read
    some amount of data, that amount should be sent through the channel, at
    which point the io.Reader will "burstily" read until it has exhausted
    the number of bytes it was told to read.

    Caution is recommended when implementing a Limiter if this bursty
    behavior is undesireable. If undesireable, make sure that any large
    ByteCounts are broken up into smaller values sent at shorter intervals.
    See BasicReader for a good example of how this can be achieved.

func BasicLimiter(b ByteCount, t time.Duration) Limiter
    BasicLimiter will divvy up the bytes into 100 smaller parts to spread
    the load across time


