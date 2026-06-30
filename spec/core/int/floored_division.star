# spec: spec.md#integers

# The // operator is floored division: the result is rounded toward
# negative infinity.
assert.eq(7 // 2, 3)
assert.eq(-7 // 2, -4)
assert.eq(7 // -2, -4)
assert.eq(-7 // -2, 3)
assert.eq(0 // 5, 0)

# The % operator yields the remainder of floored division: the result
# has the same sign as the divisor.
assert.eq(7 % 2, 1)
assert.eq(-7 % 2, 1)
assert.eq(7 % -2, -1)
assert.eq(-7 % -2, -1)

# x == (x // y) * y + (x % y) for every nonzero y.
def check_identity(x, y):
    assert.eq(x, (x // y) * y + x % y)

check_identity(17, 5)
check_identity(-17, 5)
check_identity(17, -5)
check_identity(-17, -5)

# Floored division of int operands yields an int, never a float.
assert.eq(type(7 // 2), "int")
assert.eq(type(7 % 2), "int")
