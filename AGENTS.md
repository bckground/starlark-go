# starlark-go Contributing Guide

## Project Overview

Starlark-Go is a Go implementation of Starlark, a Python-like configuration language. The project provides a complete interpreter with scanner, parser, compiler, and bytecode VM that can be embedded in Go applications.

**Import path:** `go.starlark.net/starlark`

## Commands

### Building and Testing

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./starlark
go test ./syntax
go test ./resolve

# Run a single test
go test ./starlark -run TestExecFile

# Install the starlark command-line interpreter
go install go.starlark.net/cmd/starlark@latest

# Run the interpreter
starlark                    # REPL mode
starlark script.star        # Execute file
starlark -c 'print(1+1)'    # Execute string
```

### Running .star Test Files

Most tests are in `.star` files within `testdata/` directories. These use the `starlarktest` assertion library:

```bash
# Tests are run automatically by go test
go test ./starlark          # Runs starlark/testdata/*.star
go test ./syntax            # Runs syntax/testdata/*.star
```

## Architecture

The interpreter follows a 5-stage pipeline:

```
Source → Scanning → Parsing → Resolving → Compiling → Executing
         (tokens)   (AST)     (scopes)    (bytecode)   (runtime)
```

### Package Responsibilities

**syntax/** - Lexical scanner and recursive-descent parser

- `scan.go` - Tokenization with Python-style indentation handling
- `parse.go` - LL(1) parser producing AST (Stmt and Expr nodes)
- `syntax.go` - AST node definitions

**resolve/** - Static name resolution and scope analysis

- Classifies identifiers: Universal, Predeclared, Global, Local, Free
- Assigns indices to variables for O(1) array-based lookup
- Validates control flow (break/continue in loops)
- Computes closures for nested functions

**internal/compile/** - Bytecode code generation

- Converts resolved AST to ~50 opcodes (LOAD, CALL, JMP, etc.)
- Produces `Program` with bytecode, constants, line number table
- Stack-based VM instructions with varint operand encoding

**starlark/** - Core interpreter and runtime (11k+ lines)

- `value.go` - `Value` interface and built-in types (Int, String, List, Dict, etc.)
- `eval.go` - API entry points: `ExecFile()`, `Call()`, `Eval()`
- `interp.go` - Bytecode VM execution with operand stack, locals, iterators
- `library.go` - Built-in functions (len, range, list, dict, etc.)

**repl/** - Interactive Read-Eval-Print Loop

**cmd/starlark/** - Command-line interpreter with profiling support

**lib/** - Optional standard library modules (json, math, time, proto)

**starlarkstruct/** - Struct and Module types for creating records

**starlarktest/** - Testing utilities with `assert.eq()`, `assert.error()`, etc.

### Key Design Patterns

**Value Interface:** All Starlark values implement:

```go
type Value interface {
    String() string          // String representation
    Type() string            // Type name ("int", "list", etc.)
    Freeze()                 // Make immutable for thread safety
    Truth() Bool             // Truthiness
    Hash() (uint32, error)   // For dict keys/set members
}
```

Optional interfaces add capabilities: `Callable`, `Iterable`, `Indexable`, `Mapping`, `HasAttrs`, `HasBinary`, etc.

**Thread Model:** Each `Thread` has independent execution state. Threads can share frozen (immutable) values. The `Thread.Load` callback enables custom module loading with caching.

**Error Handling:**

- Parse errors → `SyntaxError` from syntax package
- Resolution errors → returned from `resolve.File()`
- Runtime errors → `EvalError` with full backtrace

**Two-Phase Name Resolution:** Forward references are allowed within a scope. The resolver makes two passes:

1. Collect all bindings and uses
2. Match uses to bindings, compute closures

## Type Annotations Extension

This implementation includes optional static typing, a port of starlark-rust's
type annotation system (see TYPES.md for the full description):

- **Syntax**: `def f(x: int, *args: str, **kwargs: int)! -> list[int]:`,
  `x: int = 5`. With this fork's `!` error marker, `!` comes first, then `->`;
  the return annotation describes the success value (error returns skip the
  check). Lambdas cannot be annotated. Type expressions are a restricted
  grammar (paths, `T[...]`, unions with `|`, tuples, `...`); string literals
  are banned.
- **Gating**: `syntax.FileOptions.Types` is one of `TypesDisabled` (default,
  parse error), `TypesParseOnly`, `TypesEnabled`. Tests: `# option:types` /
  `# option:typesparseonly`. CLI: `starlark -types=on`, `-typecheck`.
  Positional-only parameters (`def f(x, /)`) are separately gated by
  `FileOptions.PositionalOnly` (`# option:positionalonly`, `-positionalonly`).
- **Runtime semantics**: annotations are ordinary expressions evaluated at
  `def` time in the enclosing scope (like defaults, carried in the MAKEFUNC
  tuple), converted via `starlark.TypeOf` into `*starlark.Type` matchers.
  Arguments are checked after `setArgs`, returns in the RETURN opcode,
  annotated assignments via the TYPECHECK opcode. Mismatches are uncatchable
  failures (like `fail()`), not recoverable errors. Container matching is
  deep. `list[int]`, `int | None` are first-class type values; `isinstance`
  and `eval_type` are universal builtins; `lib/typing` provides
  `typing.Any/Never/Callable/Iterable`.
- **Extensibility**: Go types implement `starlark.TypeMatcher` or
  `starlark.TypeName` to act as annotations; `starlark.Exportable` values
  learn the name of the global they're first assigned to.
- **record/enum**: opt-in packages `starlarkrecord` (record/field) and
  `starlarkenum` (enum), per starlark-rust's library extensions.
- **Static typechecker**: optional `typecheck` package
  (`typecheck.Check(file, env, loads)`) performs fixpoint inference over the
  resolved AST with intersection-based compatibility, deliberately lenient
  (unknowns become Any plus an Approximation). A module-level partial
  evaluation pre-pass (`peval.go`) tracks the type values bindings *denote*
  (aliases under conditionals, inside functions, across loads via
  `Interface.Denoted`) and lets `TypeFactory` env hooks mint nominal
  `CustomTy` types for `record`/`enum` (adapters in `starlarkrecord/typed`,
  `starlarkenum/typed`) and `error_tags` (built into `UniverseEnv`). Its
  annotation interpretation must agree with the runtime's
  (`TestAnnotationAgreement` and `TestRecordEnumAgreement` pin this); the
  universe signature table in `typecheck/universe.go` must be kept in sync
  with `starlark/library.go`.
- **Tests**: `starlark/testdata/types.star`, `types_parseonly.star`,
  parser/resolver/serialization tests, `typecheck` package tests (including a
  no-crash corpus sweep over all testdata).

## Zig-Style Error Handling Extension

This implementation includes a Zig-inspired error handling system that extends standard Starlark with explicit error propagation and handling constructs.

### Language Features

**Error-Returning Functions** - Mark functions that can return errors with `!`:

```python
def may_fail()!:
    return errors.DatabaseError
```

**Error Tag Sets** - Create namespaces containing error values:

```python
errors = error_tags("DatabaseError", "NetworkError", "TimeoutError")
# Use as: errors.DatabaseError, errors.NetworkError, etc.
```

**Try Keyword** - Explicitly propagate errors up the call stack:

```python
def caller()!:
    result = try may_fail()  # Propagates error if may_fail fails
    return result
```

**Catch Value Form** - Provide fallback values on error:

```python
result = may_fail() catch "default_value"
```

**Catch Block Form** - Execute statements on error with bound error variable:

```python
result = may_fail() catch e:
    print("Error:", e)
    recover "fallback"  # MUST end with recover or return
```

**Errdefer** - Defer execution only on error paths:

```python
def transaction()!:
    conn = connect()
    defer disconnect(conn)      # Always runs
    errdefer rollback()         # Only runs on error
    return try commit()
```

**Recover** - Resume normal execution from catch block:

```python
x = may_fail() catch e:
    log_error(e)
    recover "default"  # Exits catch block, assigns "default" to x
```

### Key Semantics

**Compile-Time Validation:**

- Calling `!` functions without try/catch is a resolver error, EXCEPT as the
  top-level call of a `defer`/`errdefer` statement (an error returned by a
  deferred call is ignored at runtime, like Go discards a deferred call's return
  values). An error-returning call passed as an _argument_ to a deferred call is
  evaluated eagerly and still requires try/catch.
- `recover` outside catch blocks is a resolver error
- `errdefer` in non-`!` functions is a resolver error

**Strictness of the compile-time check:** the "must be handled with try/catch"
rule is only enforced for calls whose target is _statically resolvable_ to a
known `!` function or error-returning builtin — i.e. a direct call by name. Calls
through a value whose error-returning-ness cannot be determined at resolve time
(a variable, parameter, `obj.method()`, `x[i]()`, or a `load()`-ed symbol) are
**not** rejected by the resolver and are instead validated at runtime. So the
static guarantee is real but partial: it catches direct misuse, not dynamic
dispatch. Treat it as a lint that covers the common case, not a soundness proof.

**Catch Block Scoping:**

- Catch blocks introduce a new lexical scope
- The error variable and variables assigned in the block are local to it, shadowing enclosing bindings of the same names without overwriting them
- Enclosing values can still be read and mutated from inside the block
- Catch blocks MUST end with either `return` or `recover` (runtime error otherwise)

**Error Propagation Model:**

- `try` checks if an error occurred and propagates it up the call stack
- Catch blocks intercept errors before propagation
- `recover` clears the error state and resumes normal execution
- Errors are represented as Error values (not Go errors)

**Errors vs failures (this duality is intentional):**

There are two distinct mechanisms, and the difference is deliberate, not an
accident of implementation:

1. **Errors — what `!` functions and error-returning builtins return.** These
   flow through a frame-local error register (`frame.pendingError`) and are
   **recoverable**: `try` propagates them, `catch` intercepts them, and `recover`
   resumes from them. Use them for expected, handleable outcomes — the Zig-style
   explicit error channel. When a `!` function returns an error that no Starlark
   caller caught and control returns to Go — whether to the embedder at the
   outermost frame or to a Go builtin that invoked the function via `Call` —
   `Call` surfaces it on its error result as a `*ReturnedError` (recoverable via
   `errors.As`; inspect `.Value`), keeping it distinct from a failure. The error
   is only ever transferred onto the caller's `frame.pendingError` when the caller
   is a Starlark function, which has the `try`/`catch` opcodes to consume it; a Go
   caller, having none, receives it explicitly instead of having it stranded on a
   frame it cannot read.

2. **Failures — unrecoverable aborts, surfaced as `*EvalError`.** A failure is
   either **explicit** (a call to `fail()`) or **implicit** (a runtime fault such
   as `1 // 0` or an arity mismatch). Failures are **uncatchable**: `try`/`catch`
   cannot intercept them, they unwind the entire stack, and execution aborts. Use
   `fail()` for programmer errors and unrecoverable conditions — the Starlark
   equivalent of a panic. (`defer`/`errdefer` still run during the unwind; a
   failure raised by cleanup is recorded on the primary `EvalError`'s `Cleanup`
   list and reported after it, each with its own backtrace, rather than masking it.)

Rule of thumb: `fail()` is for "this should never happen, stop now"; returning an
error value (e.g. `return errors.NotFound`) is for "this can happen, the caller
should decide." An error returned by a function called in a `defer`/`errdefer` is
discarded, mirroring how Go ignores a deferred call's return values.

### Implementation Architecture

The error handling system follows the same 5-stage pipeline as `defer`:

1. **Scanner** - Added tokens: CATCH, ERRDEFER, RECOVER (TRY reuses existing != tokenization)
2. **Parser** - New AST nodes: TryExpr, CatchExpr, ErrDeferStmt, RecoverStmt, DefStmt.Exclaim
3. **Resolver** - Validates error handling usage, tracks `!` functions, enforces compile-time rules
4. **Compiler** - New opcodes: TRY, CATCH_CHECK, LOAD_ERROR, ERRDEFER, RECOVER
5. **Interpreter** - Recoverable errors propagate via a frame-local error
   register (`frame.pendingError`), transferred to the caller's frame at the call
   boundary; failures propagate as Go errors that unwind the stack. Separate
   errdefer stack. Because the error register lives on the frame, it cannot leak
   across calls or executions.

### Examples

**Basic Error Handling:**

```python
errors = error_tags("NotFound")

def find_user(id)!:
    if id < 0:
        return errors.NotFound
    return "user_" + str(id)

# Value form
user = find_user(42) catch "guest"

# Block form
user = find_user(-1) catch e:
    print("Failed to find user:", e)
    recover "guest"
```

**Transaction Pattern:**

```python
def database_transaction()!:
    conn = try connect()
    defer close(conn)

    tx = try begin_transaction(conn)
    errdefer rollback(tx)

    try execute_query(tx, "INSERT ...")
    try execute_query(tx, "UPDATE ...")

    try commit(tx)
    return "success"

result = database_transaction() catch e:
    log_error("Transaction failed:", e)
    recover "failed"
```

### Known Limitations

- **Dynamic calls**: Runtime validation for function values stored in variables
- **Error variable shadowing**: Error variables in catch blocks overwrite outer variables with the same name
- **No error type checking**: All errors are string-like values, no compile-time type safety
- **Builtin support**: Go builtins must manually set `CanReturnError=true` to participate

### Testing

Test files demonstrating all features are in `starlark/testdata/`:

- `error_tags.star` - Error tag set creation and usage
- `try_propagate.star` - Error propagation chains
- `catch_blocks.star` - Catch block error handling
- `catch_scope.star` - Variable scoping in catch blocks
- `errdefer.star` - Conditional deferred cleanup
- `recover.star` - Resuming execution after errors

## Spec Suite

`spec/` holds an executable, implementation-independent specification
of Starlark and this fork's extensions, in the style of ruby/spec.
It is written in Starlark only — no Go may live under or be imported
from `spec/`. The contract between the suite and any implementation
that runs it is `spec/harness.md`; read it before adding spec files.

**Layout.** `spec/core/` covers the core language (its normative
document is `doc/spec_original.md`); `spec/optional/<unit>/` holds
one directory per independently adoptable *unit* — dialect options
(`set`, `while`, `recursion`, `toplevelcontrol`, `globalreassign`,
`positionalonly`) and extensions (`defer`, `error_handling`,
`types`) — each with a normative `spec.md` and a coverage `index.md`;
`spec/interactions/<a>+<b>/` (alphabetical, `+`-joined) holds
behavior defined only when several units combine.

**File conventions.** A spec file is one Starlark program with a
`# spec: spec.md#anchor` header naming the heading it exercises, and
optionally `# requires: unit, ...` for units beyond those implied by
its location. Files named `*_errors.star` (and only those) are
chunked: `---` separates independent programs, and `### "regex"`
(a Go-quoted string) on a line declares that the chunk fails at that
line. The runner predeclares `assert`, `trap`, `matches`, and
`freeze`; assertion failures are reported, not raised. Remember that
core files run with *default* dialect options: no top-level
control flow, no global reassignment, no `set`.

**Runner.** `spectest/` is this implementation's runner
(`go test ./spectest`); one subtest per file, gated on
`spectest.Supported`, executed in a fresh module per program.
`spectest/known_failures.txt` lists files expected to fail (reported
as skips; a listed file that passes is an error). Go's test cache
*is* `.star`-aware — it content-hashes the files the test opens and
the directory listings it walks (see `go help test`), so edits and
added/removed spec files invalidate it and a `(cached)` result is
trustworthy; `-count=1` is not needed.

**Discipline.** Spec files assert *normative* behavior, not whatever
the implementation happens to do — testdata can encode bugs (bytes
indexing did). Error-message regexes pin this implementation's
wording but are advisory for other implementations; *that* an error
occurs is the normative part. Implementation-specific behavior
(embedder modules, machine-int boundaries, resource limits) belongs
in package testdata, not in `spec/`; embedding-boundary obligations
go in prose in the unit's `spec.md`.

## Integration Entry Points

### Execute a Starlark file from Go

```go
thread := &starlark.Thread{Name: "my_program"}
globals, err := starlark.ExecFile(thread, filename, src, predeclared)
if err != nil {
    // Handle EvalError, SyntaxError, etc.
}
```

### Call a Starlark function from Go

```go
result, err := starlark.Call(thread, fn, args, kwargs)
```

### Provide custom built-in functions

```go
builtins := starlark.StringDict{
    "my_func": starlark.NewBuiltin("my_func", func(
        thread *starlark.Thread,
        b *starlark.Builtin,
        args starlark.Tuple,
        kwargs []starlark.Tuple,
    ) (starlark.Value, error) {
        // Implementation
        return starlark.None, nil
    }),
}
globals, err := starlark.ExecFile(thread, filename, src, builtins)
```

### Implement custom Starlark types in Go

Implement the `Value` interface and optional interfaces as needed:

```go
type MyType struct { /* fields */ }

func (m *MyType) String() string { return "mytype(...)" }
func (m *MyType) Type() string { return "mytype" }
func (m *MyType) Freeze() { /* make immutable */ }
func (m *MyType) Truth() starlark.Bool { return starlark.True }
func (m *MyType) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable") }

// Optional: Add methods/attributes
func (m *MyType) Attr(name string) (starlark.Value, error) { /* ... */ }

// Optional: Add operators
func (m *MyType) Binary(op syntax.Token, y starlark.Value, side starlark.Side) (starlark.Value, error) {
    /* ... */
}
```

## Important Conventions

**Concurrency:** Starlark values are mutable by default. Call `Freeze()` to make values immutable for safe sharing across goroutines. Freezing is recursive.

**Performance:**

- Global/local variable access is O(1) via indexed arrays (not hash maps)
- Dictionary iteration order is preserved (insertion order)
- No JIT compilation - pure bytecode interpretation
- Arbitrary precision integers use `math/big`

**Test File Format:** Test files (`.star`) use the `starlarktest` module:

```python
load("assert.star", "assert")

assert.eq(1 + 1, 2)
assert.true(x > 0)
assert.error(lambda: 1/0, "division by zero")
```

**Language Feature Flags:** The parser accepts optional features via `syntax.FileOptions`:

- `Set` - enable set data type
- `While` / `Recursion` - enable while loops and recursion
- `GlobalReassign` - allow multiple bindings to top-level names
- `TopLevelControl` - allow if/for/while at module level

Tests can enable these with comments like `# option:recursion`

## Debugging

Enable bytecode disassembly:

```go
import "go.starlark.net/internal/compile"
compile.Disassemble = true
```

Inspect at runtime:

```go
thread.CallStack()           // Get call stack
globals.Keys()               // List module variables
evalErr.Backtrace()          // Formatted error with stack trace
```
