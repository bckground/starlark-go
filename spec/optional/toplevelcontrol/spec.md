# Unit: toplevelcontrol

This unit permits `if`, `for`, and (with the `while` unit) `while`
statements at module level, which the core specification
(spec.md#statements) restricts to function bodies.

## Top-level control flow

Control statements at module level execute as part of module
initialization, in textual order. Names bound inside a top-level
control statement are module-level bindings.
