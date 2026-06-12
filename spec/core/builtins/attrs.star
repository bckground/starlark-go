# spec: spec.md#built-in-constants-and-functions

# dir lists the attributes (methods) of a value, sorted.
assert.contains(dir(""), "split")
assert.contains(dir([]), "append")
assert.contains(dir({}), "items")
assert.eq(dir(None), [])

# getattr retrieves an attribute by name; a default avoids the error
# for an unknown attribute.
assert.eq(getattr("a,b", "split")(","), ["a", "b"])
assert.eq(getattr("", "nope", 42), 42)
assert.fails(lambda: getattr("", "nope"), "no .nope field or method")

# hasattr reports whether a value has the named attribute.
assert.true(hasattr("", "split"))
assert.true(not hasattr("", "nope"))
assert.true(hasattr({}, "get"))

# Bound methods are first-class values.
append = [].append
assert.eq(type(append), "builtin_function_or_method")
