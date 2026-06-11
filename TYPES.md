# Type Annotations

This implementation supports optional static typing, a port of the type
annotation system of [starlark-rust](https://github.com/facebook/starlark-rust/blob/main/docs/types.md)
(used by Buck2), following that spec as closely as possible.

## Enabling

Annotations are gated by `syntax.FileOptions.Types`, mirroring
starlark-rust's `DialectTypes`:

| Mode | Behavior |
|---|---|
| `syntax.TypesDisabled` (default) | annotation syntax is a parse error |
| `syntax.TypesParseOnly` | annotations parse and are validated, but are ignored at runtime (names within them are not even resolved) |
| `syntax.TypesEnabled` | annotations parse, are validated, and are checked at runtime |

The `starlark` command exposes this as `-types=off|parse|on`, plus
`-typecheck` to run the static checker before execution. Test files
use `# option:types` / `# option:typesparseonly`.

## Syntax

```python
def fib(i: int) -> int:                  # parameters and return type
    ...

def f(x: int = 3, *args: int, **kwargs: str) -> list[int]:
    ...                                  # *args: element type; **kwargs: value type

x: int = 5                               # annotated assignment (single targets only)

def find(id: int)! -> str:               # with this fork's error marker:
    ...                                  # `!` first, then the arrow; the annotation
                                         # describes the success value
```

Lambdas cannot be annotated (the colon would be ambiguous).

A type expression is syntactically an ordinary expression, but is
restricted by the parser to: identifier paths (`int`, `typing.Any`),
parameterization (`list[int]`, `dict[str, int]`, `tuple[int, str]`,
`tuple[int, ...]`, `set[str]`, `typing.Callable[[int], str]`,
`typing.Iterable[int]`), unions (`int | None`), parenthesized tuples,
`...`, and the legacy union list `[int, str]`. Everything else —
including string literals — is a parse error.

## Semantics

Type annotations are ordinary expressions evaluated in the enclosing
scope when the `def` statement executes (exactly like parameter
defaults), then compiled into matchers stored on the function value.

- On each call, argument values (including values from defaults) are
  checked: mismatch is an uncatchable failure (`*EvalError`):
  ``Value `"a"` of type `string` does not match the type annotation `int` for argument `x` ``
- On each return (including implicit `return None`), the result is
  checked — except when a `!` function returns an error value, which
  flows through the error channel instead.
- `x: T = e` checks `e` each time the statement executes.
- Container matches are deep: `isinstance([1, "a"], list[int])` is False.

In this fork, type mismatches are **failures** (like `fail()`), not
recoverable errors: `try`/`catch` cannot intercept them. They indicate
programmer errors.

## Types as values

Types are first-class: `list[int]`, `int | None` evaluate to values of
type `type`. The universe gains `isinstance(v, t)` and `eval_type(t)`,
and the optional `typing` module (`go.starlark.net/lib/typing`)
provides `Any`, `Never`, `Callable`, and `Iterable`:

```go
predeclared := starlark.StringDict{"typing": typing.Module}
```

## Go API

- `starlark.TypeOf(v Value) (*Type, error)` converts a value in
  annotation position to a `*starlark.Type`; `(*Type).Matches(v)`
  checks a value against it.
- `(*Function).ParamType(i)` and `(*Function).ReturnType()` expose a
  function's compiled annotations.
- Embedder types implement `starlark.TypeMatcher` (custom matching) or
  `starlark.TypeName` (match by `Type()` string) to act as annotations.
- `starlark.Exportable` values are told the name of the global they
  are first assigned to (used by record/enum types).

## record and enum

Ports of starlark-rust's record/enum library extensions, as opt-in
packages (`go.starlark.net/starlarkrecord`, `go.starlark.net/starlarkenum`):

```python
Rec = record(host=str, port=field(int, 80))
r = Rec(host="localhost")     # fields type-checked at construction
r.port                        # 80

Color = enum("red", "green", "blue")
Color("red").index            # 0
Color[1]                      # Color("green")

def serve(r: Rec, c: Color):  # both work as annotations,
    ...                       # matching instances of that exact type
```

Add them via `starlarkrecord.Predeclared` / `starlarkenum.Predeclared`.

## Static typechecker

`go.starlark.net/typecheck` is an optional static analysis pass — the
analogue of starlark-rust's `AstModule.typecheck`. It runs on a
resolved file, entirely outside execution:

```go
f, _, err := starlark.SourceProgramOptions(opts, filename, src, predeclared.Has)
res, err := typecheck.Check(f, typecheck.UniverseEnv(), loads)
for _, e := range res.Errors { ... }     // e.g. "Expected type `int` but got `str`"
res.Interface                            // global types, feeds dependents' loads
res.Types.Lookup(ident)                  // per-binding inferred types
```

It infers binding types by fixpoint over assignment sites (including
`x.append(e)`-style list mutations and tuple destructuring), validates
calls against signatures (annotated defs and a hand-written table for
the universe and string/list/dict/set methods), and checks returns and
annotated assignments. It is deliberately lenient: anything it cannot
model becomes `typing.Any` plus a recorded `Approximation` — it aims
never to reject a program that would run. Compatibility is by type
*intersection*, not subtyping, exactly like starlark-rust.

Fork-specific constructs are understood: `try e` has the type of `e`;
`e catch v` is the union of both; a block-form catch unions in its
`recover` values, and the error variable types as `error`.
