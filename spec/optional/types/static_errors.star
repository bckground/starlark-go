# spec: spec.md#static-validation

# String literals are not permitted in type expressions.
def f(x: "int"): ### "string literal expression is not allowed in type expression"
    pass
---
def f() -> "int": ### "string literal expression is not allowed in type expression"
    pass
---
# Lambdas cannot be annotated. A lambda parameter annotation does not
# parse as part of the lambda.
f = lambda x: int: x ### "got ':', want newline"
