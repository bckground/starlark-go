# Type Annotations: Possible Follow-ups

Known improvement opportunities deferred from the type annotation
system (see TYPES.md). None affect correctness; they are precision or
fidelity refinements. Completed items are removed; what remains is the
work with real design prerequisites, plus deferred items.

(Done in earlier rounds: per-load typechecking in `cmd/starlark`,
golden-file test harness, pointless-comparison lint, deep-matching
cost documentation, argument-aware universe signatures, annotated-
assignment matcher caching, callable parameter-spec intersection.)

## Ongoing

- **Broaden the universe signature table.**
  *Importance: high. Effort: incremental, never "done".*
  `typecheck/universe.go` covers the universal builtins and
  string/list/dict/set methods, with argument-aware results for
  `sorted`, `min`/`max`, `zip`, `enumerate`, `list`/`set`/`tuple`/
  `dict`, `abs`, `reversed`, and `dict.get`/`pop`/`setdefault`.
  Precision can still grow (e.g. `int()` result narrowing, `str.format`
  argument checking). Anything missing types as `Any`, so additions are
  precision-only. Keep the sync test with `starlark/library.go` green.

## Big rocks (high value, real design work; do in this order)

- **Module-level partial evaluation.**
  *Importance: high — the keystone of this tier. Effort: high.*
  Alias detection (`IntList = list[int]`) is name-based and
  module-level only. rust's `fill_types_for_lint` partially evaluates
  module globals, supporting e.g. aliases under conditionals and typed
  `record(...)` constructors as annotation values. Related: aliases
  defined inside functions are not visible to annotations in nested
  defs. Prerequisite for the two items below. Constraint: degrade to
  `Any` when unsure — never reject a program that runs.

- **record/enum awareness.**
  *Importance: high. Effort: high (blocked on partial evaluation +
  a `CustomTy` API decision).*
  The static checker types `record(...)` and `enum(...)` results as
  `Any`. A `CustomTy`-style extension (mirroring rust's `TyUser`) would
  let record fields and enum elements typecheck statically, and could
  also cover `starlarkstruct`, `lib/time`, and `lib/json` values.
  Clean shape: `typecheck` defines the interface;
  `starlarkrecord`/`starlarkenum` implement it; the module-level pass
  recognizes the constructors.

- **Error-value typing.**
  *Importance: medium. Effort: low once the above exists.*
  The fork's `error` values type as an opaque prim with `Any`
  attributes. Modeling `error_tags(...)` results as a struct-like type
  with known tag members would catch typos like `errors.NotFonud`
  statically. Falls out of the same `CustomTy` + partial-evaluation
  machinery as record/enum.

## Deferred

- **REPL support for `-typecheck`.**
  *Importance: low. Effort: medium.*
  The flag only applies to file execution; the REPL path runs without
  static checking. Mechanically easy, but checking line-by-line means
  threading accumulated global types across REPL inputs — an
  incremental-Check entry point the checker wasn't designed for.

- **Lambda bodies.**
  *Importance: low. Effort: high.*
  Lambdas type as `AnyCallable` and are not descended into (rust
  parity). Inferring lambda parameter/result types from context would
  catch more errors, but rust deliberately doesn't do this — there is
  no spec to port, and contextual typing carries real false-positive
  risk. Revisit if/when rust does it.

- **LSP/editor integration.**
  *Importance: high long-term. Effort: very high — a separate project.*
  `Result.Types.Lookup(ident)` exposes per-binding inferred types; a
  language-server hover/diagnostics integration is a natural consumer.
  Everything needed (positions, per-ident types) already exists; the
  work is the server itself, which doesn't belong in this repo.
