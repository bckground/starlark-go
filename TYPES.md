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
`-typecheck` to run the static checker before execution. `-typecheck`
follows `load()` statements, checking each dependency once and feeding
its inferred module interface to its dependents, so cross-module calls
are checked precisely. Test files use `# option:types` /
`# option:typesparseonly`.

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

Positional-only parameters (`def f(x, /, y)`, Python semantics: `x`
cannot be passed by name, and its name remains available to
`**kwargs`) are supported behind a separate gate,
`syntax.FileOptions.PositionalOnly` (`starlark -positionalonly`,
`# option:positionalonly` in tests), since standard Starlark lacks
them; they compose freely with annotations.

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
- A `float` annotation also accepts ints (numeric coercion, like
  starlark-rust); the reverse does not hold.

In this fork, type mismatches are **failures** (like `fail()`), not
recoverable errors: `try`/`catch` cannot intercept them. They indicate
programmer errors.

**Performance.** Container matching is deep, so checking an annotated
parameter costs O(n) in the size of the container passed to it, on
every call — the same cost starlark-rust pays. There is deliberately
no depth or size limit (that would diverge from the spec); if profiles
show hot annotated call paths, remove the annotation or widen it to
`list` / `typing.Any`. Annotated assignments whose type expression
names only universal or predeclared values (e.g. `x: list[int] = e`,
but not a locally-defined alias) evaluate the annotation once per
function value and cache the matcher, so re-execution in loops is
cheap.

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
The `typed` subpackages of each (`starlarkrecord/typed`,
`starlarkenum/typed`) contribute matching static types to a
`typecheck.Env`, so the static checker understands them too (see
below).

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

Like starlark-rust, it reports `==`/`!=` between certainly-disjoint
types ("pointless comparisons"), and checks function values passed to
`typing.Callable[[...], R]` annotations against their actual
signatures (parameter-spec intersection). Builtin results are
argument-aware where it matters: `sorted(xs)`, `min`/`max`, `zip`,
`enumerate`, `dict.get` and friends propagate their arguments' element
types.

Fork-specific constructs are understood: `try e` has the type of `e`;
`e catch v` is the union of both; a block-form catch unions in its
`recover` values, and the error variable types as `error`. In a `!`
function, the return annotation describes the success value, so
returns also accept `error` and `error_tag` values (the runtime skips
the check for error returns).

### Module-level partial evaluation

The checker distinguishes the *type of a binding* (`IntList: type`)
from the *type value it denotes* (`list[int]`). A pre-pass — the
analogue of starlark-rust's `fill_types_for_lint` — maps each binding
to the static value its assignments compute, so annotations can refer
to computed values:

```python
T = int if legacy else int    # branches agree: T denotes int
load("lib.star", "IntList")   # aliases flow across load()

def f():
    Row = list[int]           # aliases inside functions
    def g(r: Row): ...
```

Evaluation is flow-insensitive with agree-or-unknown merging: if all
assignments to a binding compute the same static value, the binding
has it; any disagreement (or a rebinding form the evaluator does not
model) degrades it to unknown — `typing.Any` plus an `Approximation`,
never a guess. `Interface.Denoted` exposes the denoted types of a
module's globals, which is how aliases (and record/enum types) work
across `load()` with the per-module CLI flow.

### Custom types: record, enum, error_tags

A `typecheck.CustomTy` (the analogue of starlark-rust's `TyUser`) is
an externally defined static type: a display name plus an attribute
table, with optional callability, indexing, and iteration. Identity
is nominal, mirroring the runtime's pointer-identity matchers: two
structurally identical record types do not intersect.

Custom types are minted by a `typecheck.TypeFactory` attached to an
environment entry (`typecheck.WithFactory`): when a module-level
assignment calls such a builtin with statically evaluated arguments,
the partial evaluator asks the factory for the binding's static
value. Three factories exist today:

- `error_tags("NotFound", ...)` (built into `UniverseEnv`): the tag
  set's attribute table is exactly the tags, so `errors.NotFonud` is
  a static error.
- `record(host=str, port=field(int, 80))`
  (`starlarkrecord/typed.AddTypes`): unknown fields, constructor
  argument errors, and `x: Rec` annotation precision.
- `enum("red", "green")` (`starlarkenum/typed.AddTypes`): unknown
  elements (`Color.bleu`), constructor and element typing.

`TestRecordEnumAgreement` pins that the runtime `TypeMatcher` of a
record or enum type and its static `CustomTy` accept the same values
and display the same way, as `TestAnnotationAgreement` does for
annotations.
