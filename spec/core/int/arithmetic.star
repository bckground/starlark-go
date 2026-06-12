# spec: spec.md#integers

# Integers have arbitrary precision: arithmetic never overflows.
assert.eq(type(1), "int")
big = 1 << 100
assert.eq(big, 1267650600228229401496703205376)
assert.eq(big * big, 1 << 200)
assert.eq(big - big, 0)
assert.eq((1 << 64) + 1 - (1 << 64), 1)

# Basic arithmetic.
assert.eq(2 + 3, 5)
assert.eq(2 - 3, -1)
assert.eq(2 * 3, 6)
assert.eq(-(-7), 7)
assert.eq(+7, 7)

# The / operator is real division and always yields a float;
# integer division is the // operator.
assert.eq(7 / 2, 3.5)
assert.eq(type(4 / 2), "float")

# Bitwise operators.
assert.eq(5 & 3, 1)
assert.eq(5 | 3, 7)
assert.eq(5 ^ 3, 6)
assert.eq(~0, -1)
assert.eq(~5, -6)
assert.eq(1 << 10, 1024)
assert.eq(1024 >> 3, 128)
assert.eq(-16 >> 2, -4)

# Comparisons.
assert.lt(-1, 0)
assert.lt(0, 1)
assert.true(2 >= 2)
assert.true(1 != 2)

# There is no ** exponentiation operator in Starlark.
# (See parse_errors.star under core/syntax.)
