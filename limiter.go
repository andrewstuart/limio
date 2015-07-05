package limio

//A Limiter is an interface that meters some underlying discretely quantifiable
//operation with respect to time.
//
//The Limit() function, when implemented, should apply a limit to some
//underlying operation when called. Supporting concurrency is up to the
//implementer and as such, should be documented. The semantics of the channel
//are that of a token bucket. The actual integer sent through the channel
//represents a quantity of operations that can take place. The implementation
//should be sure to specify its interpretation of the quantity.
//
//Limit() returns a new boolean channel, used to comunicate that the given
//`chan int` is no longer being used and may be closed.  A false value
//indicates that the Limiter has not been shut down and may still be acted
//upon. True indicates that the limiter has been shutdown and any further
//function calls will have no effect.
//
//Unlimit() removes any formerly imposed limits and allows the underlying
//operation.
type Limiter interface {
	Limit(chan int) <-chan bool //The channel is useful for knowing that the channel has been unlimited. The boolean represents finality.
	Unlimit()
}
