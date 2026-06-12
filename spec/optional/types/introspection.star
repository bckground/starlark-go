# spec: spec.md#isinstance-and-eval_type

# isinstance matches values against types, deeply.
assert.true(isinstance(1, int))
assert.true(isinstance("s", str))
assert.true(isinstance([1, 2], list))
assert.true(isinstance([1, 2], list[int]))
assert.true(not isinstance([1, "a"], list[int]))
assert.true(isinstance(None, None))
assert.true(isinstance(1, int | None))
assert.true(isinstance(None, int | None))
assert.true(not isinstance("s", int | None))
assert.true(isinstance({"a": [1]}, dict[str, list[int]]))
assert.true(not isinstance(1, str))

# eval_type converts a value denoting a type into a type value.
T = eval_type(int | None)
assert.eq(type(T), "type")
assert.true(isinstance(1, T))

# A value that does not denote a type is rejected.
assert.fails(lambda: eval_type(5), "value of type int is not a type")
