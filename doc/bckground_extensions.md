# Starlark Extensions

This document describes the language extensions added to this Starlark
implementation beyond the
[standard specification](spec.md).  These extensions are specific to
the Go implementation at `go.starlark.net/starlark`.

  * [Defer statements](#defer-statements)
  * [Error handling](#error-handling)
    * [Data types](#data-types)
    * [Expressions](#expressions)
    * [Statements](#statements)
    * [Built-in functions](#built-in-functions)
    * [Static validation](#static-validation)
    * [Module-level try](#module-level-try)
    * [Execution model](#execution-model)
    * [Examples](#examples)


## Defer statements

A `defer` statement schedules a function call to be executed when the
enclosing function returns.

```grammar {.good}
DeferStmt = 'defer' CallExpr .
```

The operand of `defer` must be a function call expression.  The
function and all its arguments are evaluated immediately when the
`defer` statement is executed, but the call itself is not performed
until the enclosing function returns.

```python
def process():
    f = open("data.txt")
    defer close(f)
    return parse(f)    # close(f) runs after parse completes
```

A `defer` statement may appear only inside a function body.  It is a
static error to use `defer` at module level.

### Multiple defers

A function may contain multiple `defer` statements.  Deferred calls
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

### Argument evaluation

Arguments to a deferred call are evaluated at the point of the `defer`
statement, not at the point where the deferred call executes.  This is
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

### Execution guarantees

Deferred calls execute on all exit paths of the enclosing function:
normal return, early return, and error propagation (see
[Error handling](#error-handling) below).  This makes `defer` suitable
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

### Nested functions

Each function has its own independent defer stack.  Deferred calls in
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

### Keyword arguments

Deferred calls support the same argument-passing conventions as
ordinary calls, including positional arguments, keyword arguments, and
mixed forms.

```python
def f():
    defer log(level="info", msg="done")
    work()
```

---


## Error handling

This section describes the error handling extension, inspired by Zig's
error return traces.  It extends the language with explicit error
values, error propagation, and structured error handling, giving
programs a way to represent, propagate, and recover from expected
failure conditions without resorting to sentinel values or the `fail`
built-in.

Standard Starlark provides no mechanism by which errors can be handled
within the language (see [spec.md](spec.md#module-execution)).

### Overview

The error handling system introduces the following constructs:

- **Error-returning functions** (`def f()!:`) -- functions that may return an error value.
- **Error tag sets** (`error_tags(...)`) -- namespaces of error tag values.
- **`try`** -- explicit error propagation from a `!` call to the enclosing `!` function.
- **`catch`** -- interception and handling of errors from `!` calls.
- **`errdefer`** -- deferred calls that execute only on error paths.
- **`recover`** -- resumption of normal execution from within a `catch` block.

The central design principle is that error flow is always visible in
the source code: a call to an error-returning function must be wrapped
in `try` or `catch`, and error propagation never happens silently.


### Data types

#### Error tags

An _error tag_ is an immutable, hashable value that identifies a class
of error.  Its [type](#type) is `"error_tag"`.

Error tags are created by the built-in function `error_tags`
(see [Built-in functions](#error_tags)).  They are compared by
identity: two tags are equal only if they are the same object.

An error tag used in a Boolean context is considered false.

```python
errors = error_tags("NotFound", "Timeout")
type(errors.NotFound)           # "error_tag"
bool(errors.NotFound)           # False
errors.NotFound == errors.NotFound  # True
errors.NotFound == errors.Timeout   # False
```

#### Errors

An _error_ is an immutable value that pairs an error tag with optional
metadata.  Its [type](#type) is `"error"`.

An error has the following attributes:

```text
error.tag        the error tag (error_tag)
error.message    a human-readable description (string)
error.cause      an underlying error, or None (error or None)
error.details    supplementary data (list)
```

An error used in a Boolean context is considered false.

Errors are not created directly by the programmer.  Instead, when a `!`
function returns an error tag, the runtime wraps it in an error value
that carries the tag.  Errors with richer metadata can be constructed
by calling an error tag as a function:

```python
errors = error_tags("IOError")
e = errors.IOError(message="disk full", details=["sda1"])
e.tag       # IOError
e.message   # "disk full"
e.details   # ["sda1"]
```

#### Error tag sets

An _error tag set_ is an immutable namespace whose attributes are error
tags.  Its [type](#type) is `"error_tag_set"`.  Error tag sets are created by the
`error_tags` built-in function.


### Expressions

#### Try expressions

A `try` expression propagates an error from a call to an
error-returning function.

```grammar {.good}
TryExpr = 'try' CallExpr .
```

The operand of `try` must be a call to an error-returning function
(one defined with `!`).  If the call succeeds, `try` yields the
result.  If the call returns an error, `try` causes the enclosing
function to return immediately, propagating the error to its caller.

```python
def read_config()!:
    data = try read_file("config.json")  # propagates error
    return try parse_json(data)          # propagates error
```

`try` may appear only inside an error-returning function or at module
level:

- Inside a `!` function, `try` propagates the error to the caller.
- At module level, `try` converts the error to a `fail()` call,
  terminating execution.

It is a static error to use `try` inside a function that is not
marked with `!`, or to use `try` on a call to a function that is
not marked with `!`.

```python
def f():
    x = try may_fail()     # static error: try requires enclosing error-returning function

def g()!:
    x = try normal()       # static error: try requires call to error-returning function
```

#### Catch expressions

A `catch` expression intercepts an error from a call to an
error-returning function.  It has two forms: _value form_ and
_block form_.

```grammar {.good}
CatchExpr = CallExpr 'catch' Test .                              # value form
CatchExpr = CallExpr 'catch' identifier ':' Suite .              # block form
```

**Value form.**
If the call succeeds, the catch expression yields the call's result.
If the call returns an error, the catch expression yields the value
of the fallback expression instead.

```python
config = read_config() catch default_config
port = read_port() catch 8080
```

**Block form.**
If the call succeeds, the catch expression yields the call's result
and the block is not executed.
If the call returns an error, the error is bound to the named variable
and the block is executed.  The block must end with either a `return`
statement or a `recover` statement; failing to do so is a dynamic error.

```python
config = read_config() catch e:
    print("config error:", e)
    recover default_config
```

A catch expression may appear in any function (not only `!` functions)
and at module level.  It is the primary way for non-`!` code to call
`!` functions.

It is a static error to use `catch` on a call to a function that is
not marked with `!`.

**Scoping.**
Catch blocks do not create a new lexical scope; they behave like `if`
statement bodies.  Variables assigned in a catch block, including the
error variable, are visible in the enclosing scope.


### Statements

#### Error-returning function definitions

A function definition may include a `!` marker after the parameter
list to indicate that the function can return error values.

```grammar {.good}
DefStmt = 'def' identifier '(' [Parameters [',']] ')' ['!'] ':' Suite .
```

An error-returning function is like an ordinary function, except:

1. When it returns an error tag or error value, the runtime converts the
   return into an error signal that is visible to `try` and `catch` in
   the caller.  The function appears to return `None` to the caller;
   the error is carried out of band.

2. It may contain `try` expressions and `errdefer` statements.

```python
def connect(host, port)!:
    sock = try open_socket(host, port)
    defer close(sock)
    return try handshake(sock)
```

It is a static error for a non-`!` function to contain a `try`
expression or an `errdefer` statement.

Every call to a `!` function must be wrapped in `try` or `catch`.
It is a static error to call a `!` function without either.

#### Errdefer statements

An `errdefer` statement registers a deferred function call that
executes only if the enclosing function exits with an error.

```grammar {.good}
ErrDeferStmt = 'errdefer' CallExpr .
```

An `errdefer` statement may appear only inside an error-returning
function.  It is a static error to use `errdefer` in a non-`!`
function.

Like `defer`, arguments to the deferred call are evaluated
immediately, but the call itself is deferred.  Multiple `errdefer`
statements execute in LIFO (last-in, first-out) order, and they
execute before any regular `defer` statements.

If the function returns successfully (no error), errdeferred calls
are discarded without execution.

```python
def process_file(path)!:
    f = try open(path)
    defer close(f)
    errdefer log("failed to process " + path)

    data = try read(f)
    return try transform(data)

# If transform fails:  log() runs, then close() runs.
# If transform succeeds: only close() runs.
```

#### Recover statements

A `recover` statement may appear only inside the block form of a
`catch` expression.  It ends execution of the catch block and provides
the value that the catch expression yields.

```grammar {.good}
RecoverStmt = 'recover' Test .
```

```python
result = may_fail() catch e:
    log_error(e)
    recover "default"    # catch expression yields "default"

# result is either the successful return value or "default"
```

It is a static error to use `recover` outside of a catch block.

A catch block that does not end with `recover` or `return` causes a
dynamic error at runtime.


### Built-in functions

#### error_tags

`error_tags(*names)` creates an error tag set containing one error tag for
each of the given string arguments.  The tags are accessible as
attributes of the returned error tag set.

Each error tag is a callable: calling it with optional keyword
arguments `message`, `cause`, and `details` produces an error value
with richer metadata.

```python
errors = error_tags("NotFound", "PermissionDenied", "Timeout")

errors.NotFound             # an error tag
type(errors.NotFound)       # "error_tag"

# Create an error with metadata:
e = errors.NotFound(message="user 42 not found")
e.tag                       # NotFound
e.message                   # "user 42 not found"
```


### Static validation

The following rules are enforced at compile time (during name resolution):

1. **`defer` requires a function body.**
   A `defer` statement may appear only inside a function.
   Using `defer` at module level is a static error.

2. **`try` requires a `!` call.**
   The operand of `try` must be a direct call to a function defined with `!`.

3. **`catch` requires a `!` call.**
   The operand of `catch` must be a direct call to a function defined with `!`.

4. **`try` requires a `!` enclosure or module level.**
   `try` may appear only inside a `!` function or at module level.
   Using `try` inside a non-`!` function is a static error.

5. **Every `!` call must be handled.**
   A call to a `!` function must appear as the operand of `try` or `catch`.
   A bare call to a `!` function, even inside another `!` function, is a
   static error.

6. **`errdefer` requires a `!` function.**
   An `errdefer` statement may appear only inside a `!` function.
   Using `errdefer` in a non-`!` function, or at module level, is a
   static error.

7. **`recover` requires a `catch` block.**
   A `recover` statement may appear only inside the block form of a `catch`
   expression.


### Module-level try

At module level, `try` does not propagate errors (there is no caller
to propagate to).  Instead, it is compiled as a `catch` that calls the
`fail` built-in with the error value, terminating module execution.

```python
# Module level:
config = try load_config()
# Equivalent to:
# config = load_config() catch e:
#     fail(e)
```

When `fail` is called with an error value, Go callers can use
`errors.As` with `*starlark.FailError` to extract the underlying
Starlark error:

```go
var failErr *starlark.FailError
if errors.As(err, &failErr) {
    fmt.Println(failErr.Value.Tag)  // the error tag
}
```


### Execution model

#### Error propagation

When an error-returning function executes a `return` statement whose
value is an error tag or error value, the following happens:

1. The error value is stored as the _pending error_ on the thread.
2. The function's result is set to `None`.
3. Errdeferred calls execute in LIFO order.
4. Regular deferred calls execute in LIFO order.
5. The function returns `None` to its caller.

The caller then observes the pending error through `try` or `catch`:

- **`try`**: checks for a pending error; if present, runs the current
  function's own errdefers, then returns `None` with the error still
  pending.
- **`catch`**: checks for a pending error; if present, materializes
  the error value, clears the pending error, and executes the fallback.

If neither `try` nor `catch` handles the error, static validation has
already rejected the program.

#### Deferred call ordering

When a function exits with an error, the execution order is:

1. Errdeferred calls (LIFO order)
2. Regular deferred calls (LIFO order)

When a function exits successfully:

1. Errdeferred calls are discarded
2. Regular deferred calls (LIFO order)


### Examples

#### Resource cleanup with defer

```python
def process_files(paths):
    results = []
    for path in paths:
        f = open(path)
        defer close(f)
        results.append(parse(f))
    return results
```

#### Basic error handling

```python
errors = error_tags("NotFound", "InvalidInput")

def find_user(id)!:
    if id < 0:
        return errors.InvalidInput
    if id > 1000:
        return errors.NotFound
    return "user_" + str(id)

# Value form: provide a default
user = find_user(-1) catch "guest"      # "guest"

# Block form: inspect the error
user = find_user(9999) catch e:
    if e.tag == errors.NotFound:
        recover "anonymous"
    else:
        recover "error: " + str(e)
```

#### Error propagation chain

```python
errors = error_tags("DBError", "ServiceError")

def query_db(sql)!:
    if "DROP" in sql:
        return errors.DBError
    return [{"id": 1}]

def get_user(id)!:
    rows = try query_db("SELECT * WHERE id=" + str(id))
    if len(rows) == 0:
        return errors.ServiceError
    return rows[0]

def handle_request()!:
    user = try get_user(42)
    return "Hello, " + str(user)

# Catch at the boundary
response = handle_request() catch e:
    recover "Error: " + str(e)
```

#### Transaction pattern with errdefer

```python
errors = error_tags("CommitFailed")

def transaction(db)!:
    tx = try begin(db)
    defer close(tx)
    errdefer rollback(tx)       # only if we error out

    try execute(tx, "INSERT INTO users VALUES (1, 'alice')")
    try execute(tx, "INSERT INTO logs VALUES (1, 'created')")
    try commit(tx)
    return "success"

result = transaction(db) catch e:
    recover "transaction failed: " + str(e)
```

#### Nested error handling

```python
errors = error_tags("ParseError", "ValidationError")

def parse(input)!:
    if not input:
        return errors.ParseError
    return input.split(",")

def validate(fields)!:
    if len(fields) < 3:
        return errors.ValidationError
    return fields

def process(input)!:
    fields = try parse(input)
    valid = try validate(fields)
    return valid

# Outer code catches everything
result = process("a,b,c") catch e:
    recover []

# Or handle specific stages differently
def process_with_recovery(input)!:
    fields = parse(input) catch e:
        recover ["default"]
    return try validate(fields)
```
