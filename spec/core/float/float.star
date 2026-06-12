# spec: spec.md#floating-point-numbers

# Floating-point numbers use IEEE 754 double precision.
assert.eq(type(1.0), "float")
assert.eq(type(1e3), "float")
assert.eq(1e3, 1000.0)
assert.eq(0.25 + 0.25, 0.5)
assert.eq(1.5 * 2, 3.0)

# Mixed int/float arithmetic yields float.
assert.eq(type(1 + 0.5), "float")
assert.eq(1 + 0.5, 1.5)

# An int and a float compare equal if they denote the same value.
assert.eq(1.0, 1)
assert.eq(0.0, 0)
assert.lt(1, 1.5)
assert.lt(1.5, 2)

# Division.
assert.eq(7.0 / 2, 3.5)
assert.eq(7.0 // 2, 3.0)
assert.eq(type(7.0 // 2), "float")
assert.eq(-7.0 // 2, -4.0)
assert.eq(7.0 % 2, 1.0)
assert.fails(lambda: 1.0 / 0, "floating-point division by zero")
assert.fails(lambda: 1 / 0.0, "floating-point division by zero")
assert.fails(lambda: 1.0 // 0, "floored division by zero")
assert.fails(lambda: 1.0 % 0, "floating-point modulo by zero")

# float(x) converts a number or a decimal string.
assert.eq(float(3), 3.0)
assert.eq(float("1.5"), 1.5)
assert.eq(float(True), 1.0)
assert.fails(lambda: float("pi"), "invalid float literal")

# Unlike IEEE 754, comparison defines a total order on floats:
# NaN compares equal to itself and greater than +infinity.
nan = float("nan")
inf = float("inf")
assert.eq(nan, nan)
assert.lt(inf, nan)
assert.lt(1e308, inf)
assert.lt(-inf, 0)
