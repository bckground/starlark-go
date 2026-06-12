# Unit: while

This unit adds the `while` statement, which the core specification
(spec.md#while-loops) makes optional.

## While loops

```
WhileStmt = 'while' Test ':' Suite .
```

A `while` statement evaluates its condition; while the condition is
true, the loop body executes. `break` and `continue` behave as in
`for` loops. Like all control statements, `while` may appear only
within a function unless the `toplevelcontrol` unit is also
supported.
