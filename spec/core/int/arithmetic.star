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

# Division and remainder sign rules hold at any magnitude.
def check_division():
    for m in [1, (1 << 31) - 1]:
        assert.eq((100 * m) // (7 * m), 14)
        assert.eq((100 * m) // (-7 * m), -15)
        assert.eq((-100 * m) // (7 * m), -15)
        assert.eq((-100 * m) // (-7 * m), 14)
        assert.eq((100 * m) % (7 * m), 2 * m)
        assert.eq((100 * m) % (-7 * m), -5 * m)
        assert.eq((-100 * m) % (7 * m), 5 * m)
        assert.eq((-100 * m) % (-7 * m), -2 * m)

check_division()

# Multiplication is exact at any magnitude.
assert.eq(0x1000000000000001 * 0x1000000000000001,
          0x1000000000000002000000000000001)
assert.eq(1111111111111111 * 1111111111111111,
          1234567901234567654320987654321)

# A negative shift count is an error.
assert.fails(lambda: 2 << -1, "negative shift count")
assert.fails(lambda: 2 >> -1, "negative shift count")

# | and ^ have distinct precedence: ^ binds tighter.
assert.eq(1 | 0 ^ 1, 1)

# There is no ** exponentiation operator in Starlark.
# (See parse_errors.star under core/syntax.)
