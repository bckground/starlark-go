# spec: spec.md#dictionaries

# get returns the value or a default (None if unspecified).
d = {"a": 1}
assert.eq(d.get("a"), 1)
assert.eq(d.get("z"), None)
assert.eq(d.get("z", 0), 0)

# keys, values, and items return new lists in insertion order.
d2 = {"x": 1, "y": 2}
assert.eq(d2.keys(), ["x", "y"])
assert.eq(d2.values(), [1, 2])
assert.eq(d2.items(), [("x", 1), ("y", 2)])

# setdefault returns the value, first inserting the default if the
# key is absent.
d3 = {"a": 1}
assert.eq(d3.setdefault("a", 0), 1)
assert.eq(d3.setdefault("b", 2), 2)
assert.eq(d3, {"a": 1, "b": 2})
assert.eq(d3.setdefault("c"), None)

# pop removes and returns the value; a default avoids the error for a
# missing key.
d4 = {"a": 1}
assert.eq(d4.pop("a"), 1)
assert.eq(d4, {})
assert.eq(d4.pop("a", 0), 0)
assert.fails(lambda: d4.pop("a"), "pop: missing key")

# popitem removes and returns the first item.
d5 = {"a": 1, "b": 2}
assert.eq(d5.popitem(), ("a", 1))
assert.eq(d5, {"b": 2})
assert.fails(lambda: {}.popitem(), "empty dict")

# update inserts from pairs and keyword arguments.
d6 = {"a": 1}
d6.update([("b", 2)], c=3)
d6.update({"a": 0})
assert.eq(d6, {"a": 0, "b": 2, "c": 3})

# clear removes all items.
d7 = {"a": 1}
d7.clear()
assert.eq(d7, {})
