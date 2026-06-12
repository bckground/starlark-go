# Starlark Spec Suite: Harness Contract

This document defines the contract between the spec suite and any
implementation that runs it: the layout of the suite, the format of
spec files, the assertion vocabulary an implementation must provide,
and the obligations of a conforming runner.

The suite is written in Starlark and is implementation-independent.
It contains no Go (or Rust, or Java); an implementation participates
by providing the small harness described here. The reference runner
for `go.starlark.net` lives outside this directory — nothing under
`spec/` may import or depend on it.

## Layout

```
spec/
  harness.md            this document
  core/                 the core language (doc/spec-original.md)
  optional/<unit>/      one directory per optional unit
  interactions/<a>+<b>/ behavior requiring two or more units
```

A **unit** is an independently adoptable body of language behavior
with its own normative specification. Core dialect options that the
base specification already treats as optional (`set`, `while`,
`recursion`, `toplevelcontrol`, `globalreassign`) are units, as are
language extensions (`defer`, `error_handling`, `types`,
`positionalonly`). Each `optional/<unit>/` directory contains:

- `spec.md` — the normative specification of the unit.
- `index.md` — a map from `spec.md` headings to the spec files that
  exercise them, so coverage gaps are visible.
- the spec files.

Interaction directories (`interactions/error_handling+types/`) hold
files whose behavior is defined only when all the named units are
supported; they are named by the participating units, joined with
`+`, in alphabetical order.

## Spec files

A spec file is an ordinary Starlark program. By default a file
contains **one** program: the runner executes it as a single module
in a fresh environment.

Files whose name ends in `_errors.star` are **chunked**: `---` lines
divide the file into chunks, each an independent program executed in
its own fresh environment, and lines containing `###` declare
expected failures (see "Aborts" below). Chunking is permitted only in
`*_errors.star` files; everything else is one assertion program per
file.

### Headers

A spec file begins with comment directives:

```python
# requires: error_handling, defer
# spec: spec.md#errdefer
```

- `# requires:` — the units this file needs, beyond those implied by
  its location. A file under `optional/<unit>/` implicitly requires
  that unit; a file under `interactions/<a>+<b>/` implicitly requires
  all the named units; a file under `core/` implicitly requires
  nothing. A runner must skip (not fail) any file requiring a unit
  the implementation does not support.
- `# spec:` — the heading of the normative document that this file
  exercises. Every file must carry one; `index.md` is derived from
  these anchors.

## Assertion vocabulary

The runner predeclares the following names in every spec program.
Assertion failures are **reported, not raised**: a failed assertion
records a test failure and execution of the program continues, so one
file may report several failures in a single run.

The `assert` module:

| Name | Behavior |
|---|---|
| `assert.eq(x, y)` | reports a failure unless `x == y` |
| `assert.ne(x, y)` | reports a failure unless `x != y` |
| `assert.true(cond, msg="assertion failed")` | reports `msg` unless `cond` is true |
| `assert.lt(x, y)` | reports a failure unless `x < y` |
| `assert.contains(x, y)` | reports a failure unless `y in x` |
| `assert.fails(f, pattern)` | calls `f()`; reports a failure unless the call aborts with a message matching the regular expression `pattern` |
| `assert.fail(msg)` | reports `msg` unconditionally |

Free functions:

| Name | Behavior |
|---|---|
| `trap(f)` | calls `f()`; returns the abort message as a string, or `None` if the call completed normally |
| `matches(pattern, str)` | reports whether `str` matches the regular expression `pattern` |
| `freeze(x)` | freezes `x` and everything reachable from it, returning `x`; used to spec behavior of frozen values |

Regular expressions use RE2/`re`-style syntax restricted to the
common subset: literal text, character classes, `.`, `*`, `+`, `?`,
alternation, and anchors. Spec files must not rely on
engine-specific features.

`trap` is the primitive; `assert.fails` is the preferred surface and
spec files should use it unless they need to branch on success or
inspect the message. Implementations may define both in terms of
whatever error representation they have, provided the observable
behavior above holds.

## Observation channels

Spec files observe behavior through three channels:

1. **Values.** Expressions and `assert.*` over their results. The
   bulk of the suite.

2. **Recoverable errors** (units that define them, e.g.
   `error_handling`). Observed *in the language*, using the unit's
   own constructs — a `try`/`catch` spec catches the error it is
   specifying. No harness support is involved; this is deliberate, as
   it keeps the unit self-hosting.

3. **Aborts** — errors that terminate execution and that no language
   construct can intercept. In core Starlark every dynamic error is
   an abort; in the `error_handling` unit, aborts are the "failures"
   half of its errors/failures distinction. Two mechanisms:

   - `assert.fails(f, pattern)` / `trap(f)`, when the aborting
     computation can be wrapped in a callable.
   - **Chunk expectations**, when it cannot — parse errors, resolve
     (static) errors, and aborts at module top level. In a
     `*_errors.star` file, `### "regex"` on a line declares that
     executing the chunk fails *at that line* with a message matching
     the regex (a Go-style quoted string):

     ```python
     x = 1 // 0 ### "division by zero"
     ---
     def f():
         recover 1 ### "recover outside catch block"
     ```

     A chunk with no `###` expectation must execute without error.

## Normative force of error assertions

*That* an error occurs is normative: a conforming implementation must
abort (or must not) exactly where the suite says.

The error *message* regex is advisory: it pins the reference
implementation's wording and exists to keep assertions honest, but a
foreign implementation with different wording is not non-conforming.
A runner for such an implementation may relax message matching to
occurrence-only, or carry a local overlay of message patterns.
Error *positions* (line/column) are asserted only as far as the chunk
convention implies (the line carrying `###`) and only for static
errors; dynamic-error positions are not part of the contract.

## Runner obligations

A conforming runner must:

1. **Discover** every `*.star` file under the suite root.
2. **Gate**: compute each file's required units (location + header)
   and skip files whose requirements the implementation does not
   meet. Skips must be reported as skips, not silently dropped and
   not counted as passes.
3. **Isolate**: execute each program (file, or chunk of a
   `*_errors.star` file) as a fresh module — fresh globals, fresh
   thread/evaluation state. Nothing carries over between programs.
4. **Predeclare** the assertion vocabulary above, configured so that
   assertion failures are reported to the host test framework without
   halting the program.
5. **Match expectations**: for chunked files, verify the set of
   actual errors against the `###` expectations (position and, per
   the implementation's chosen strictness, message); for plain files,
   any abort is a test failure.
6. **Report** per-file pass/fail/skip and exit non-zero if any
   gated-in file fails.

A runner may additionally support a **known-failures overlay**: an
out-of-tree list of spec files expected to fail for that
implementation, reported distinctly (like ruby/spec's tags). The
overlay lives with the implementation, never in this directory.

## Scope

The suite specifies behavior observable from Starlark programs.
Embedding-boundary behavior — host-language error types, freezing
across threads, the API by which host functions participate in a
unit — is specified prose-only in each unit's `spec.md`
("implementation obligations") and tested by each implementation's
own test suite, not here.
