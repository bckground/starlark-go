# spec: spec.md#lexical-elements

# Comparison operators do not chain.
x = 1 < 2 < 3 ### `< does not associate with <`
---
# There is no ** exponentiation operator.
x = 2 ** 10 ### "got '\\*\\*', want"
---
# A positional argument may not follow a named argument.
def f(a, b):
    pass

y = f(a=1, 2) ### "positional argument may not follow named"
---
# Parameter names must be unique.
def g(x, x): ### "duplicate parameter: x"
    pass
---
# Indentation must be consistent.
def h():
    x = 1
   y = 2 ### "unindent does not match any outer indentation level"
---
# Integer literals may not have redundant leading zeros.
z = 08 ### "invalid int literal"
