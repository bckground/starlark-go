# Unit: error_handling

This unit adds an explicit error channel to Starlark, in the style of
Zig: functions declare that they can return errors, callers must
visibly propagate or handle them, and unhandled-able faults remain
aborts. It introduces the binary `!` marker on function definitions,
the `try` and `catch` expression forms, the `recover` and `errdefer`
statements, the `error_tags` built-in function, and the `error_tag`
and `error` data types.

The `errdefer` statement is defined only when the `defer` unit is
also supported.

## Overview

The unit distinguishes two failure channels:

- **Errors** are values returned by error-returning functions. They
  are expected, recoverable outcomes: `try` propagates them, `catch`
  intercepts them, `recover` resumes from them.
- **Failures** are aborts: `fail(...)` and implicit faults such as
  division by zero. No construct of this unit intercepts them; they
  terminate execution as in core Starlark.

Rule of thumb: a failure means "this should never happen"; an error
means "this can happen, and the caller should decide".

## Error tags

`error_tags(*names)` returns an *error tag set* (type `"error_tags"`)
with one attribute per name, each a distinct *error tag* (type
`"error_tag"`).

```python
errors = error_tags("NotFound", "Timeout")
errors.NotFound             # an error_tag
str(errors.NotFound)        # "NotFound"
```

Each call to `error_tags` mints fresh tags: tags are equal only to
themselves, so two sets' tags differ even under the same name.
Accessing an undeclared tag name is an error. Two tag sets may be
merged with `|` or `+`, yielding a new set containing the tags of
both (the tags themselves, not copies).

## Error values

An error tag is callable. Calling it with the optional keyword
arguments `message`, `cause`, and `extra` produces an *error value*
(type `"error"`) carrying those fields:

```python
e = errors.NotFound(message="user 42 not found", extra={"id": 42})
e.tag        # errors.NotFound
e.message    # "user 42 not found"
e.cause      # an underlying error, or None
e.extra      # caller-supplied context, or None
```

When an error-returning function returns a bare tag, the error value
observed by the handler has that tag and the tag's name as its
message. Handlers discriminate errors by comparing `e.tag` against
tags.

## Error-returning functions

```
DefStmt = 'def' identifier '(' [Parameters] ')' ['!'] ':' Suite .
```

The `!` marker after the parameter list declares that a function may
return an error. Within such a function, a `return` statement whose
value is an error tag or error value returns it on the error channel:
the caller observes the error, not a normal result. Returning any
other value is an ordinary return. Lambda expressions cannot carry
the marker.

A call to an error-returning function is itself an expression that
may produce an error; it must be handled with `try` or `catch` (see
Static validation).

## try

```
TryExpr = 'try' Test .
```

`try x` evaluates the call `x`; if it produced an error, the
enclosing error-returning function returns that error to its own
caller (after running its errdefers); otherwise the value of `try x`
is the call's result. `try` is permitted only inside error-returning
functions, except at module level.

At module level, `try x` is instead equivalent to handling the error
with a catch block that calls `fail` with the error value: a
module-level error that reaches `try` becomes an abort.

## catch

```
CatchExpr  = Test 'catch' Test .
CatchBlock = Test 'catch' identifier ':' Suite .
```

In the *value form*, `f() catch v` evaluates to the call's result, or
to the value `v` if the call produced an error. `v` is evaluated only
on the error path. The two forms are distinguished after the `catch`
keyword: an identifier immediately followed by `:` selects the block
form; anything else is a value-form fallback expression. (In the rare
position where a value-form fallback ending in an identifier abuts a
`:` — for example inside a slice — parenthesize the fallback.)

In the *block form*, an error binds the error value to the
identifier and executes the suite, which must terminate with `return`
or `recover`; falling off the end of a catch block is a failure
("catch block must end with recover or return"). The catch block
introduces a new lexical scope: the error variable and any names
bound inside the block are local to it, shadowing enclosing bindings
of the same names and leaving them unchanged. The block may read and
mutate enclosing values; results escape through `recover`'s operand,
`return`, or mutation.

Both forms require the operand to be a call to an error-returning
function. Both forms are permitted in any function and at module
level.

## recover

```
RecoverStmt = 'recover' Test .
```

`recover` is permitted only inside a catch block. It ends the block:
the catch expression's result is the operand's value and execution
resumes normally after the catch block.

## errdefer

*Requires the `defer` unit.*

```
ErrDeferStmt = 'errdefer' CallExpr .
```

`errdefer` schedules a call like `defer`, but the call runs only if
the enclosing function exits with an error. It is a static error
outside an error-returning function. On an error exit, errdeferred
calls run first (LIFO), then deferred calls (LIFO); on a successful
exit, errdeferred calls are discarded.

## Failures

`try`, `catch`, and `recover` operate only on the error channel.
A failure — `fail(...)` or an implicit fault — is not intercepted by
`catch` and propagates as an abort exactly as in core Starlark.
Deferred calls still run during the unwind. An error returned by a
function called by a `defer` or `errdefer` statement itself is
discarded, like the call's return value.

## Static validation

The following are static errors:

- a direct call to an error-returning function (or error-returning
  built-in) that is not the operand of `try` or `catch`;
- `try` or `catch` applied to a call that cannot return an error;
- `try` outside an error-returning function (except at module level);
- `recover` outside a catch block;
- `errdefer` outside a function, or in a function without the `!`
  marker.

The first check applies only to calls whose target is statically
resolvable to a known error-returning function — a direct call by
name. Calls through values (variables, parameters, methods, elements)
are validated at runtime instead: a dynamic call that misuses the
error channel is a failure.

## Implementation obligations

This section is prose-only; conformance is tested by each
implementation's own suite, not by spec files.

- An embedder's native functions must be able to participate as
  error-returning callees, and calls to them must be subject to the
  same handling requirements, at least dynamically.
- An error returned by an error-returning function that no enclosing
  handler catches and that reaches the embedding boundary must be
  surfaced to the embedder distinctly from a failure, preserving the
  error value.
- Error state must be confined to the call in which it arises; it
  must not leak across threads, modules, or independent executions.
