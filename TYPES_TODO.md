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

The three items below are one project in three stages: teach the
static checker that *values computed at module load time* can carry
type information. Today it only understands type information that is
syntactically visible (annotations and the name-based alias map).
The first item is the engine, the second the vocabulary, the third a
client that falls out of both.

It is tempting to special-case `record(...)` recognition without the
general evaluator, but every such shortcut re-creates the alias map's
problems one level up — order sensitivity, no conditionals, no
cross-module flow. With the evaluator first, record/enum/error_tags
are each just an env hook, and the leniency story stays in one place.
A reasonable first slice is `error_tags` (see below): it exercises
the evaluator design without needing `CustomTy` construction from
type arguments.

Cross-cutting obligation: `TestAnnotationAgreement` pins that the
static interpretation of annotations matches the runtime's. Stages 2
and 3 add new annotation-position values (record types, enum types),
so that test must grow alongside — the runtime `TypeMatcher` for a
record and its static `CustomTy` must accept the same values.

- **Module-level partial evaluation.**
  *Importance: high — the keystone of this tier. Effort: high
  (3–5 days); the risk is in threading resolution, not the evaluator.*
  `collectAliases` does one linear scan for the literal shape
  `Name = <type-expression>`, so all of these fail today: aliases
  under conditionals (`T = int if legacy else float`), aliases defined
  inside a function and used by a nested def's annotation, aliases
  loaded from another module, and any alias whose right-hand side is a
  *call* — which is exactly what `record(...)`, `enum(...)`, and
  `error_tags(...)` are.
  The fix is rust's `fill_types_for_lint` shape: a small abstract
  interpreter over top-level statements that maps each global binding
  to a *static value* — "denotes the type `list[int]`", "denotes this
  record type", or unknown — handling assignment, if/else (merge
  branches; disagreement → unknown), and loads. The key conceptual
  change: distinguish the **type of a binding** (`IntList: type`) from
  the **type value it denotes** (`list[int]`); the checker currently
  models only the former. `Interface` grows a second channel for
  denoted type values, which is what makes cross-module aliases work
  with the per-load CLI flow. Constraint: when the evaluator does not
  understand something, degrade to `Any` plus an `Approximation` —
  never guess, never reject a program that runs.

- **record/enum awareness.**
  *Importance: high. Effort: high (~1 week on top of partial
  evaluation; blocked on it plus the `CustomTy` API decision).*
  The checker needs a way to *represent* "an instance of the record
  type bound to `Rec`" — rust's `TyUser`. Ours: a `CustomTy`
  extension interface in `typecheck` (name, attribute table, call
  signature for construction, intersection rule), wrapped as a new
  `Basic` alternative. Two design points:
  - *No import cycle.* `typecheck` imports only `resolve` and
    `syntax` and should stay that way: it defines the interface, and
    adapters live with the implementations (`starlarkrecord`/
    `starlarkenum`, or `typed` subpackages), contributing `Env`
    entries. The partial evaluator, on seeing
    `Rec = record(host=str, port=field(int, 80))`, evaluates the
    arguments as type values and asks the `record` env entry's hook to
    mint a `CustomTy`.
  - *Identity is nominal.* At runtime a record type matches instances
    by pointer identity; the static analogue is the global binding
    that holds it (the `bindKeyOf` identity the solver already uses).
    Two structurally identical records must not intersect.
  What it buys — all silent today: `r.hosst` (no such attribute),
  `Rec(host=8)` (field type), `def serve(x: Rec)` precision,
  `Color.bleu` (no such enum element). The same mechanism then covers
  `starlarkstruct`, `lib/time`, and `lib/json` values.

- **Error-value typing.**
  *Importance: medium — catches the single most common runtime
  failure of the error extension. Effort: ~1 day once the above
  exist; also viable as the first slice of the evaluator.*
  The fork's `error` values type as an opaque prim with `Any`
  attributes. `errors = error_tags("NotFound", "Timeout")` is the
  easiest case of the whole pattern: a top-level call whose arguments
  are *string literals*, so the evaluator needs no type-expression
  machinery — it can mint a `CustomTy` whose attribute table is
  exactly the tag set, each member of type `error_tag`. Since
  `error_tags` is universal in this fork, the hook lives directly in
  `typecheck/universe.go`; no adapter package needed. Catches typos
  like `errors.NotFonud` statically.

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
