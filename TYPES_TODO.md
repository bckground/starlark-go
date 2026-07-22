# Type Annotations: Possible Follow-ups

Known improvement opportunities deferred from the type annotation
system (see TYPES.md). None affect correctness; they are precision or
fidelity refinements. Completed items are removed; what remains is the
work with real design prerequisites, plus deferred items.

(Done in earlier rounds: per-load typechecking in `cmd/starlark`,
golden-file test harness, pointless-comparison lint, deep-matching
cost documentation, argument-aware universe signatures, annotated-
assignment matcher caching, callable parameter-spec intersection.
Done in this round — the former "big rocks", see TYPES.md: module-
level partial evaluation with the binding-type/denoted-type split and
`Interface.Denoted`; the `CustomTy`/`TypeFactory` extension API;
error-tag-set typing in `universe.go`; record/enum awareness via the
`starlarkrecord/typed` and `starlarkenum/typed` adapters; `!`-function
returns relaxed to accept `error`/`error_tag`.)

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

## Follow-ups to the custom-type work

The three "big rocks" (module-level partial evaluation, record/enum
awareness, error-value typing) shipped; see TYPES.md for what exists.
Natural extensions, none blocking:

- **Cover `starlarkstruct`, `lib/time`, and `lib/json` values.**
  *Importance: medium. Effort: low–medium each.*
  The `CustomTy`/`TypeFactory` mechanism now exists; these are
  additional adapters. `struct(...)` is a factory like `record(...)`
  (kwargs become the attribute table, values typed by inference
  rather than denoted types); `module(...)` likewise. `lib/time` and
  `lib/json` need only static `Env` entries (moduleBasic tables or
  CustomTy values), no factories.

- **Static checking of `field()` defaults.**
  *Importance: low (the runtime checks them at load time anyway).*
  `field(int, "80")` is only caught when the module executes. The
  field factory sees the default's static value and could check
  intersection with the field type when it knows both.

- **Tag-set merges.** `a | b` on two known error tag sets currently
  degrades to `Any` (lenient); the evaluator could understand the
  merge and mint the union table.

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
