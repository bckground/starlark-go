# Defer statements

This document describes the `defer` statement, a language extension
added to this Starlark implementation beyond the
[standard specification](spec.md). It is specific to the Go
implementation at `go.starlark.net/starlark`.

- [Multiple defers](#multiple-defers)
- [Argument evaluation](#argument-evaluation)
- [Execution guarantees](#execution-guarantees)
- [Nested functions](#nested-functions)
- [Keyword arguments](#keyword-arguments)
- [Examples](#examples)

A `defer` statement schedules a function call to be executed when the
enclosing function returns.

```grammar {.good}
DeferStmt = 'defer' CallExpr .
```

The operand of `defer` must be a function call expression. The
function and all its arguments are evaluated immediately when the
`defer` statement is executed, but the call itself is not performed
until the enclosing function returns.

```python
def process():
    f = open("data.txt")
    defer close(f)
    return parse(f)    # close(f) runs after parse completes
```

A `defer` statement may appear only inside a function body. It is a
static error to use `defer` at module level.

## Multiple defers

A function may contain multiple `defer` statements. Deferred calls
execute in LIFO (last-in, first-out) order, like a stack: the most
recently deferred call runs first.

```python
def f():
    log = []
    defer log.append("first")
    defer log.append("second")
    defer log.append("third")
    log.append("body")
    return log

f()     # ["body", "third", "second", "first"]
```

## Argument evaluation

Arguments to a deferred call are evaluated at the point of the `defer`
statement, not at the point where the deferred call executes. This is
important when variables are reassigned between the `defer` and the
function's return.

```python
def f():
    result = []
    i = 0
    defer result.append(i)   # captures i=0
    i = 1
    defer result.append(i)   # captures i=1
    i = 2
    result.append(i)         # appends 2 immediately
    return result

f()     # [2, 1, 0]
```

## Execution guarantees

Deferred calls execute on all exit paths of the enclosing function:
normal return, early return, and error propagation (see
[Error handling](error_handling.md)). This makes `defer` suitable
for cleanup operations that must always run.

```python
def process(flag):
    result = []
    defer result.append("cleanup")
    if flag:
        result.append("early")
        return result       # cleanup runs here
    result.append("normal")
    return result           # cleanup runs here too
```

## Nested functions

Each function has its own independent defer stack. Deferred calls in
an inner function do not affect the outer function's defer stack, and
vice versa.

```python
def outer():
    result = []
    defer result.append("outer-defer")
    def inner():
        defer result.append("inner-defer")
        result.append("inner-body")
    inner()
    result.append("outer-body")
    return result

outer()     # ["inner-body", "inner-defer", "outer-body", "outer-defer"]
```

## Keyword arguments

Deferred calls support the same argument-passing conventions as
ordinary calls, including positional arguments, keyword arguments, and
mixed forms.

```python
def f():
    defer log(level="info", msg="done")
    work()
```

## Examples

### Resource cleanup

```python
def process_files(paths):
    results = []
    for path in paths:
        f = open(path)
        defer close(f)
        results.append(parse(f))
    return results
```
