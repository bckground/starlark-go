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

# dict() constructs a dictionary from keyword arguments or an
# iterable of key/value pairs.
assert.eq(dict(), {})
assert.eq(dict(a=1, b=2), {"a": 1, "b": 2})
assert.eq(dict([("a", 1), ("b", 2)]), {"a": 1, "b": 2})
assert.eq(dict([("a", 1)], a=2), {"a": 2})

# A frozen dict cannot be mutated.
frozen = freeze({"a": 1})

def mutate_frozen():
    frozen["b"] = 2

assert.fails(mutate_frozen, "cannot insert into frozen hash table")
