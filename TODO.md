## Move basic rate-based limiting to a Manager.
an io.Reader should only accept a Limit channel and have the management of said
channel elsewhere.

A Limiter implements the control of some transaction.
A Manager decides when/how much to limit.

### io.Reader example
limio.Reader exposes Limit() method
BasicManager uses a rate and a ticker to set the limit.

## Make it easier to implement strategies

How?
- Default loop implementation?
- Strategy type definition?
  - Would it be passed []chan int or []limiter?

## Remaining Questions
- Is a Manager interface useful (aside from a spec?) What could a function do if it accepts a Manager?
  - Calling Manage(), it could abstract the creation of other interfaces and manage them itself.
    - Like returning an io.Reader implementation that's already limited by the Manager
