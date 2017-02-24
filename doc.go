//Package limio provides an interface for rate limiting as well as a
//rate-limited Reader implementation.
//
//In limio, there are two important interfaces for rate limiting. The first,
//Limiter, is more primitive. Most times, implementers of a Limiter should be
//creating a way to apply time constraints to a discretely quantifiable
//transaction.
//
//The other interface is a Manager, which will likely be implemented in many
//more cases, as it allows consumers to take any number of Limiters and apply a
//strategy to the group. Most importantly, a Manager will also need to
//implement the Limiter interface, allowing consumers to treat its encapsulated
//group of Limiters as a single Limiter, knowing the strategy will be applied
//within the given limits.
package limio // import "astuart.co/limio"
