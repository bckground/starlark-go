# spec: spec.md#built-in-constants-and-functions

# len reports the number of elements of a string, bytes, list,
# tuple, dict, or range; other types are an error.
assert.eq(len("abc"), 3)
assert.eq(len([1, 2]), 2)
assert.eq(len((1,)), 1)
assert.eq(len({"a": 1}), 1)
assert.eq(len(range(10)), 10)
assert.fails(lambda: len(3), "value of type int has no len")
assert.fails(lambda: len(None), "value of type NoneType has no len")

# range yields a lazy integer sequence; it supports len, indexing,
# membership, and iteration.
r = range(1, 10, 3)
assert.eq(type(r), "range")
assert.eq(list(r), [1, 4, 7])
assert.eq(r[2], 7)
assert.eq(len(range(0)), 0)
assert.eq(list(range(3)), [0, 1, 2])        # one argument: stop
assert.eq(list(range(2, 5)), [2, 3, 4])     # two arguments: start, stop
assert.eq(list(range(5, 2, -1)), [5, 4, 3]) # negative step
assert.true(7 in range(10))
assert.fails(lambda: range(1, 10, 0), "step argument must not be zero")

# enumerate pairs indices with elements; an optional argument sets
# the first index.
assert.eq(list(enumerate(["a", "b"])), [(0, "a"), (1, "b")])
assert.eq(list(enumerate(["a"], 5)), [(5, "a")])

# reversed returns a new list in reverse order.
assert.eq(reversed([1, 2, 3]), [3, 2, 1])
assert.eq(reversed(range(3)), [2, 1, 0])

# sorted returns a new sorted list; the sort is stable and accepts
# key= and reverse=.
assert.eq(sorted([3, 1, 2]), [1, 2, 3])
assert.eq(sorted(["b", "a"]), ["a", "b"])
assert.eq(sorted([3, 1, 2], reverse=True), [3, 2, 1])
assert.eq(sorted(["aaa", "b", "cc"], key=len), ["b", "cc", "aaa"])
assert.fails(lambda: sorted([1, "a"]), "string < int not implemented")

# zip combines iterables element-wise, stopping at the shortest.
assert.eq(list(zip([1, 2, 3], [4, 5])), [(1, 4), (2, 5)])
assert.eq(list(zip()), [])
assert.eq(list(zip([1], [2], [3])), [(1, 2, 3)])

# max and min accept an iterable or multiple arguments, and key=.
assert.eq(max([1, 5, 2]), 5)
assert.eq(max(1, 5, 2), 5)
assert.eq(min([3, 1, 2]), 1)
assert.eq(max([1, -5], key=abs), -5)
assert.fails(lambda: max([]), "max: argument is an empty sequence")
