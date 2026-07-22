# spec: spec.md#operators

x = set([1, 2, 3])
y = set([3, 4])

# Union, intersection, difference, symmetric difference.
assert.eq(x | y, set([1, 2, 3, 4]))
assert.eq(x & y, set([3]))
assert.eq(x - y, set([1, 2]))
assert.eq(x ^ y, set([1, 2, 4]))

# Operators yield new sets; the operands are unmodified.
union = x | y
union.add(99)
assert.eq(x, set([1, 2, 3]))
assert.eq(y, set([3, 4]))
