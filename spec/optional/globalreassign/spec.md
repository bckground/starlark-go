# Unit: globalreassign

This unit permits reassignment of module-level names, which the core
specification (spec.md#name-binding-and-variables) forbids: in core
Starlark a global may be bound by at most one statement.

## Reassigning globals

With this unit, a module-level name may be bound by multiple
statements; each assignment replaces the binding. Functions that read
the global observe its value at call time. Predeclared names may also
be shadowed and rebound at module level.
