# spec: spec.md#tuples

# A tuple is an immutable sequence.
t = (1, "two", 3.0)
assert.eq(type(t), "tuple")
assert.eq(len(t), 3)

# A 1-tuple requires a trailing comma; parentheses alone group.
assert.eq(type((1,)), "tuple")
assert.eq(type((1)), "int")
assert.eq(len(()), 0)

# A tuple may be written without parentheses in some contexts.
pair = 1, 2
assert.eq(pair, (1, 2))

# Indexing and slicing.
assert.eq(t[0], 1)
assert.eq(t[-1], 3.0)
assert.eq(t[1:], ("two", 3.0))

# Concatenation and repetition.
assert.eq((1, 2) + (3,), (1, 2, 3))
assert.eq((0,) * 3, (0, 0, 0))

# Tuples are immutable.
assert.fails(lambda: t.append, "has no .append field or method")

def set_elem():
    t[0] = 9

assert.fails(set_elem, "tuple value does not support item assignment")

# Tuples compare lexicographically and are hashable (if their
# elements are).
assert.lt((1, 2), (1, 3))
assert.eq((1, (2, 3)), (1, (2, 3)))
d = {(1, 2): "x"}
assert.eq(d[(1, 2)], "x")

# Unpacking.
a, b = (1, 2)
assert.eq((b, a), (2, 1))
(c, d2), e = (1, 2), 3
assert.eq((c, d2, e), (1, 2, 3))

# tuple() converts any iterable.
assert.eq(tuple([1, 2]), (1, 2))
assert.eq(tuple(), ())
