# Type Annotations: Possible Follow-ups

Known improvement opportunities deferred from the initial implementation
of the type annotation system (see TYPES.md). None affect correctness;
they are performance, precision, or fidelity refinements.

## Runtime

- **Cache annotated-assignment matchers.** `x: T = e` re-evaluates the
  type expression `T` and re-converts it via `TypeOf` on every
  execution of the statement (spec-faithful, but slower than
  starlark-rust, which compiles a `TypeCompiled` once). A pc-keyed
  matcher cache on the TYPECHECK opcode would make repeated executions
  (e.g. in loops) near-free. Only safe when the type expression is
  side-effect-free, which the restricted grammar already guarantees.

- **`float` accepting `int` (numeric coercion).** starlark-rust's
  runtime `float` matcher accepts ints. This port is strict: an `int`
  does not match a `float` annotation (one-line change in
  `starlark/type.go` `nameType.matches`, plus the corresponding rule in
  `typecheck/oracle.go` `intersects`, if rust parity is preferred).

- **Deep container matching cost.** Checking every element of large
  lists/dicts on each annotated call is O(n) per parameter (rust pays
  the same). Consider documenting a depth or size limit, or a
  per-Thread opt-out, if profiles show hot annotated call paths.

## Static typechecker (`typecheck` package)

- **Broaden the universe signature table.** `typecheck/universe.go`
  covers the universal builtins and string/list/dict/set methods, but
  precision can grow indefinitely (e.g. `min`/`max` element types,
  `sorted` propagating its argument's element type, `zip` arity-aware
  tuples). Anything missing types as `Any`, so additions are
  precision-only. Keep the sync test with `starlark/library.go` green.

- **Callable parameter-spec intersection.** Two callables currently
  always intersect; rust's `params_intersect` compares signatures.
  Needed only for checking higher-order annotations like
  `typing.Callable[[int], str]` against actual function signatures.

- **Lambda bodies.** Lambdas type as `AnyCallable` and are not
  descended into (rust parity). Inferring lambda parameter/result types
  from context would catch more errors.

- **Module-level partial evaluation.** Alias detection
  (`IntList = list[int]`) is name-based and module-level only. rust's
  `fill_types_for_lint` partially evaluates module globals, supporting
  e.g. aliases under conditionals and typed `record(...)` constructors
  as annotation values. Related: aliases defined inside functions are
  not visible to annotations in nested defs.

- **record/enum awareness.** The static checker types `record(...)` and
  `enum(...)` results as `Any`. A `CustomTy`-style extension (mirroring
  rust's `TyUser`) would let record fields and enum elements typecheck
  statically, and could also cover `starlarkstruct`, `lib/time`, and
  `lib/json` values.

- **Golden-file test harness.** The current tests are table-driven in
  Go. A `testdata/golden/*.golden` harness with an `-update` flag
  (mirroring rust's `typing/tests/golden`) would make transliterating
  more of rust's golden corpus cheaper and diffs reviewable.

- **Pointless-comparison lint.** `x == y` on certainly-disjoint types
  currently passes (returns `bool`); rust reports it. Low-risk lint to
  add in `binaryType` for EQL/NEQ.

- **Error-value typing.** The fork's `error` values type as an opaque
  prim with `Any` attributes. Modeling `error_tags(...)` results as a
  struct-like type with known tag members would catch typos like
  `errors.NotFonud` statically.

## Tooling

- **REPL support for `-typecheck`.** The flag only applies to file
  execution; the REPL path runs without static checking.

- **Per-load typechecking in `cmd/starlark`.** `-typecheck` passes
  `loads: nil`, so symbols from `load()` type as `Any`. Wiring the
  load callback to recursively `Check` dependencies and feed their
  `Interface`s would extend checking across modules.

- **LSP/editor integration.** `Result.Types.Lookup(ident)` exposes
  per-binding inferred types; a language-server hover/diagnostics
  integration is a natural consumer.
