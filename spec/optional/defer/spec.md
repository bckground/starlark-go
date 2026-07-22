# Unit: defer

This unit adds the `defer` statement, which schedules a call to run
when the enclosing function exits.

## Defer statements

```
DeferStmt = 'defer' CallExpr .
```

The operand of `defer` must be a call expression. Executing the
statement does not perform the call; it schedules it to run when the
enclosing function exits. `defer` may appear only within a function;
at module level it is a static error.

## Argument evaluation

The callee and its arguments are evaluated when the `defer` statement
executes, not when the deferred call runs. Only the call itself is
delayed. All argument-passing conventions of ordinary calls are
permitted.

## Multiple defers

Each function call has its own defer stack. Deferred calls run in
LIFO order: the last `defer` executed is the first to run. A `defer`
inside a loop schedules one call per iteration.

## Execution guarantees

Deferred calls run on every exit path of the function: normal return,
early return, and abort. When a function exits because of an abort,
its deferred calls (and those of every function being unwound) still
run; the abort then continues.

The return value of a deferred call is discarded.

## Nested functions

A nested function has its own independent defer stack; its deferred
calls run when it exits, not when the enclosing function exits.
