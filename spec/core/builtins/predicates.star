# spec: spec.md#built-in-constants-and-functions

# any reports whether any element of the iterable is true.
assert.eq(any([0, False, ""]), False)
assert.eq(any([0, 1]), True)
assert.eq(any([]), False)

# all reports whether all elements of the iterable are true.
assert.eq(all([1, True, "x"]), True)
assert.eq(all([1, 0]), False)
assert.eq(all([]), True)

# Elements are tested for truth, not equality with True.
assert.eq(any([None, "", [1]]), True)
assert.eq(all([1, "a", [0]]), True)
