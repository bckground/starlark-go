# Unit: positionalonly

This unit adds positional-only parameters to function definitions,
following PEP 570 syntax.

## Positional-only parameters

```
Parameters = PosOnlyParams '/' ',' OtherParams | OtherParams .
```

A `/` in a parameter list marks every parameter before it as
positional-only: callers must supply those parameters positionally,
never by name. Parameters after `/` behave as usual. A positional-only
parameter may have a default value.

Passing a positional-only parameter by name is a dynamic error, as is
a `/` with no parameters before it, or more than one `/`, which are
static errors.
