Limiter

Can accept some limit. Concrete implementations likely are either Managers or
limit some primitive transaction (like an io.Reader interface).

Manager

Encapsulates any number of limiters as a group, and uses one strategy to
manage a group of Limiters.

# Goals
- Limiters can implement strategies, but *must* expose the ability to be fully
  controlled externally
- A Limiter should impose no limits until it is explicitly given one.
- Managers *should* encapsulate the management of their Limiters, and expose a
  way for the group to be Limited as a group.
- Managers *may* have their own internal token source, as long as it can be
  replaced by another at any time.

# Questions

## Should a manager be a limiter?
Yes, because then strategies can be layered as desired.
They maintain their own internal token bucket, and should not delegate control.
