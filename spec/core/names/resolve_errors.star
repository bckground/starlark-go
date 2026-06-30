# spec: spec.md#name-binding-and-variables

# A reference to a name that is bound nowhere is a static error.
x = undeclared ### "undefined: undeclared"
---
# Static errors are reported even when the code is unreachable.
def f():
    if False:
        return undeclared ### "undefined: undeclared"
