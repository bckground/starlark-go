# spec: spec.md#lists

# A list is a mutable sequence.
l = [1, "two", 3.0]
assert.eq(type(l), "list")
assert.eq(len(l), 3)
assert.eq(len([]), 0)

# Indexing, negative indexing, and element update.
assert.eq(l[0], 1)
assert.eq(l[-1], 3.0)
nums = [1, 2, 3]
nums[1] = 20
assert.eq(nums, [1, 20, 3])
assert.fails(lambda: [1, 2][5], "list index 5 out of range")

# Slicing yields a new list.
sq = [0, 1, 4, 9, 16]
assert.eq(sq[1:3], [1, 4])
assert.eq(sq[::2], [0, 4, 16])
assert.eq(sq[::-1], [16, 9, 4, 1, 0])
copy = sq[:]
copy[0] = 99
assert.eq(sq[0], 0)

# Concatenation and repetition yield new lists.
assert.eq([1, 2] + [3], [1, 2, 3])
assert.eq([0] * 3, [0, 0, 0])
assert.fails(lambda: [1] + (2,), "unknown binary op: list \\+ tuple")

# Membership.
assert.true(2 in [1, 2, 3])
assert.true(4 not in [1, 2, 3])

# Lists compare lexicographically, element by element.
assert.eq([1, [2, 3]], [1, [2, 3]])
assert.lt([1, 2], [1, 3])
assert.lt([1, 2], [1, 2, 0])

# list() converts any iterable; with no argument it is a new empty list.
assert.eq(list(), [])
assert.eq(list((1, 2)), [1, 2])
assert.eq(list({"a": 1}), ["a"])

# A frozen list cannot be mutated.
frozen = freeze([1, 2])
assert.fails(lambda: frozen.append(3), "cannot append to frozen list")

# A list cannot be mutated while being iterated.
def mutate_during_iteration():
    seq = [1, 2, 3]
    return [seq.append(x) for x in seq]

assert.fails(mutate_during_iteration, "cannot append to list during iteration")
