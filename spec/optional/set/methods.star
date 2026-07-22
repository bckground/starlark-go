# spec: spec.md#methods

# add inserts an element; inserting an existing element is a no-op.
s = set([1])
s.add(2)
s.add(1)
assert.eq(s, set([1, 2]))

# remove deletes an element; an absent element is an error.
# discard deletes if present and never fails.
s2 = set([1, 2])
s2.remove(1)
assert.eq(s2, set([2]))
assert.fails(lambda: s2.remove(9), "remove: missing key")
s2.discard(9)
s2.discard(2)
assert.eq(s2, set())

# pop removes and returns the first element.
s3 = set([7, 8])
assert.eq(s3.pop(), 7)
assert.eq(s3.pop(), 8)
assert.fails(lambda: s3.pop(), "empty set")

# clear removes all elements.
s4 = set([1, 2])
s4.clear()
assert.eq(s4, set())

# union, intersection, difference, and symmetric_difference accept
# any iterable and yield a new set.
assert.eq(set([1, 2]).union([2, 3]), set([1, 2, 3]))
assert.eq(set([1, 2]).intersection([2, 3]), set([2]))
assert.eq(set([1, 2]).difference([2]), set([1]))
assert.eq(set([1, 2]).symmetric_difference([2, 3]), set([1, 3]))

# issubset and issuperset.
assert.true(set([1, 2]).issubset([1, 2, 3]))
assert.true(not set([1, 9]).issubset([1, 2]))
assert.true(set([1, 2, 3]).issuperset([1, 2]))

# update inserts the elements of an iterable.
s5 = set([1])
s5.update([2, 3])
assert.eq(s5, set([1, 2, 3]))
