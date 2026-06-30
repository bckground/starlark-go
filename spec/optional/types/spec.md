# Unit: types

This unit adds optional runtime-checked type annotations, following
the design of starlark-rust's type system. Annotations may appear on
function parameters, return types, and assignments; values are
checked against them as the program runs, and a mismatch is a failure
(an abort), not a recoverable error.

## Annotation syntax

```
DefStmt    = 'def' identifier '(' [Parameters] ')' ['->' Type] ':' Suite .
Parameter  = identifier [':' Type] ['=' Test] .
AnnAssign  = identifier ':' Type '=' Test .
```

Parameters of every kind (positional, `*args`, keyword-only,
`**kwargs`) may carry `: T` annotations; a function may declare a
return type with `-> T`; an assignment to a single name may carry an
annotation. Lambda expressions cannot be annotated.

A type expression is a restricted expression grammar: dotted paths
(`int`, `typing.Any`), parameterized types (`list[int]`,
`dict[str, int]`), unions (`int | None`), tuples, and `...`.
String literals are not permitted in type expressions.

## Annotation semantics

Annotations are ordinary expressions, evaluated once — when the `def`
or assignment statement executes — in the enclosing scope, like
default values. The resulting value must denote a type (for example
the built-in type functions `int`, `str`, `list`, a parameterized
type such as `list[int]`, a union, or `None`); a value that does not
denote a type is a failure at that point.

- A parameter annotation is checked against each argument when the
  function is called.
- A return annotation is checked against the function's result at
  each return.
- An assignment annotation is checked against the assigned value when
  the assignment executes.

A mismatch is a failure: it cannot be intercepted by any language
construct, and it aborts execution like `fail`.

## Type values

`list[int]`, `int | None`, and similar type expressions are
first-class values (type `"type"`) and may be bound to names and used
in later annotations.

Container matching is deep: `list[int]` accepts only a list whose
every element is an int; `dict[str, int]` checks every key and value;
`tuple[int, str]` checks arity and each element.

## isinstance and eval_type

`isinstance(x, t)` reports whether the value `x` matches the type
denoted by `t`, with the same deep-matching rules as annotations.

`eval_type(t)` converts a value denoting a type into a first-class
type value, failing if the value does not denote a type.

## Static validation

The following are static errors:

- annotation syntax outside the restricted type-expression grammar;
- a string literal in a type expression;
- annotations on a lambda.

A well-formed annotation whose expression evaluates to a non-type is
a dynamic failure, reported when the `def` or assignment executes.

## Implementation obligations

This section is prose-only; conformance is tested by each
implementation's own suite, not here.

- The implementation must provide a mode in which annotations are
  parsed but not enforced, and a mode in which the syntax is
  rejected, so that the unit is adoptable incrementally.
- Native values supplied by the embedder must be able to participate
  as types in annotations.
