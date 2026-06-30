# spec: spec.md#the-set-type

# set(x) collects the distinct elements of an iterable; with no
# argument it is a new empty set.
s = set([1, 2, 2, 3])
assert.eq(type(s), "set")
assert.eq(len(s), 3)
assert.eq(len(set()), 0)
assert.eq(set((1, 2)), set([1, 2]))

# Membership and iteration; insertion order is preserved.
assert.true(2 in s)
assert.true(9 not in s)
assert.eq(list(set([3, 1, 2])), [3, 1, 2])
assert.eq([x for x in set("ab".elems())], ["a", "b"])

# Truth.
assert.eq(bool(set()), False)
assert.eq(bool(set([0])), True)

# Equality ignores order; sets are not ordered by <.
assert.eq(set([1, 2]), set([2, 1]))
assert.ne(set([1]), set([2]))

# Elements must be hashable; sets are not hashable.
assert.fails(lambda: set([[1]]), "unhashable type: list")
assert.fails(lambda: {set(): 1}, "unhashable type: set")

# A frozen set cannot be mutated.
frozen = freeze(set([1]))
assert.fails(lambda: frozen.add(2), "add: cannot insert into frozen hash table")
