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
- Calling `!` functions without try/catch is always a resolver error
- `recover` outside catch blocks is a resolver error
- `errdefer` in non-`!` functions is a resolver error

**Catch Block Scoping:**
- Catch blocks do NOT create new lexical scopes (like if statements)
- Variables assigned in catch blocks are visible outside the block
- The error variable (e.g., `e`) shadows any existing variable in the current scope
- Catch blocks MUST end with either `return` or `recover` (runtime error otherwise)

**Error Propagation Model:**
- `try` checks if an error occurred and propagates it up the call stack
- Catch blocks intercept errors before propagation
- `recover` clears the error state and resumes normal execution
- Errors are represented as Error values (not Go errors)

### Implementation Architecture

The error handling system follows the same 5-stage pipeline as `defer`:

1. **Scanner** - Added tokens: CATCH, ERRDEFER, RECOVER (TRY reuses existing != tokenization)
2. **Parser** - New AST nodes: TryExpr, CatchExpr, ErrDeferStmt, RecoverStmt, DefStmt.Exclaim
3. **Resolver** - Validates error handling usage, tracks `!` functions, enforces compile-time rules
4. **Compiler** - New opcodes: TRY, CATCH_CHECK, LOAD_ERROR, ERRDEFER, RECOVER
5. **Interpreter** - Runtime error propagation via Thread.pendingErrorValue, separate errdefer stack

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
