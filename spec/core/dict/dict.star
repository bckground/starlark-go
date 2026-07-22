# spec: spec.md#dictionaries

# A dict maps hashable keys to values.
d = {"a": 1, "b": 2}
assert.eq(type(d), "dict")
assert.eq(len(d), 2)
assert.eq(d["a"], 1)

# Lookup of a missing key is an error; use get for a default.
assert.fails(lambda: d["z"], 'key "z" not in dict')

# Insertion and update.
m = {}
m["x"] = 1
m["x"] = 2
assert.eq(m, {"x": 2})

# Keys may be of any hashable type; values are unrestricted.
mixed = {1: "int", "1": "string", (1, 2): "tuple", True: None}
assert.eq(mixed[(1, 2)], "tuple")
assert.fails(lambda: {[1]: 1}, "unhashable type: list")

def unhashable_insert():
    {}[{}] = 1

assert.fails(unhashable_insert, "unhashable type: dict")

# A dict preserves insertion order; updating an existing key does not
# affect its position.
ordered = {"b": 1, "a": 2}
ordered["c"] = 3
ordered["b"] = 9
assert.eq(list(ordered), ["b", "a", "c"])

# Membership tests keys, not values.
assert.true("a" in d)
assert.true(1 not in d)

# Iteration yields keys, in insertion order.
assert.eq([k for k in {"x": 1, "y": 2}], ["x", "y"])

# Equality is by content; dicts are not ordered.
assert.eq({"a": 1, "b": 2}, {"b": 2, "a": 1})
assert.fails(lambda: {} < {}, "dict < dict not implemented")

# Dicts are unhashable.
assert.fails(lambda: {{}: 1}, "unhashable type: dict")

# Duplicate keys are not permitted in a dict literal.
assert.fails(lambda: {"aa": 1, "bb": 2, "bb": 4}, 'duplicate key: "bb"')

# An unparenthesized tuple may be a key in element assignment.
pairs = {}
pairs[1, 2] = 3
assert.eq(pairs.keys()[0], (1, 2))

# A re-inserted key keeps its original position but takes the new
# value, even as the dict grows past rehashing.
grown = dict([("a", 0), ("b", 1), ("c", 2), ("b", 3)])
assert.eq(grown.keys(), ["a", "b", "c"])
assert.eq(grown["b"], 3)
grown.update([("d", 4), ("e", 5), ("f", 6), ("g", 7), ("h", 8), ("i", 9), ("j", 10), ("k", 11)])
assert.eq(grown.keys(), ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"])

# dict() constructs a dictionary from keyword arguments or an
# iterable of key/value pairs.
assert.eq(dict(), {})
assert.eq(dict(a=1, b=2), {"a": 1, "b": 2})
assert.eq(dict([("a", 1), ("b", 2)]), {"a": 1, "b": 2})
assert.eq(dict([("a", 1)], a=2), {"a": 2})
assert.eq(dict((["a", 2], ["a", 3]), a=4), {"a": 4})

# A keyword may collide with a positional-argument key, but not with
# another keyword.
assert.fails(lambda: dict({"b": 3}, a=4, **dict(a=5)), 'duplicate keyword arg: "a"')

# A frozen dict cannot be mutated.
frozen = freeze({"a": 1})

def mutate_frozen():
    frozen["b"] = 2

assert.fails(mutate_frozen, "cannot insert into frozen hash table")
