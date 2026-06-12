# spec: spec.md#lambda-expressions

# A lambda expression yields an anonymous function whose body is a
# single expression.
square = lambda x: x * x
assert.eq(type(square), "function")
assert.eq(square(4), 16)

# Lambdas accept the same parameter forms as def.
assert.eq((lambda: 42)(), 42)
assert.eq((lambda x, y=10: x + y)(1), 11)
assert.eq((lambda *args: args)(1, 2), (1, 2))
assert.eq((lambda **kw: kw)(a=1), {"a": 1})

# Lambdas close over enclosing variables.
def make_adder(n):
    return lambda m: n + m

assert.eq(make_adder(3)(4), 7)

# A lambda may be used where any expression is expected.
assert.eq(sorted([3, 1, 2], key=lambda x: -x), [3, 2, 1])
