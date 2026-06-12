# spec: spec.md#positional-only-parameters

# / must follow at least one parameter.
def f(/, x): ### "/ must follow at least one parameter"
    pass
---
# At most one / is allowed.
def f(a, /, b, /): ### "multiple / parameters not allowed"
    pass
