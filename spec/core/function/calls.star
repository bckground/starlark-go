# spec: spec.md#function-and-method-calls

def f(a, b, c=3):
    return (a, b, c)

# Arguments bind by position, by name, or both.
assert.eq(f(1, 2), (1, 2, 3))
assert.eq(f(1, b=2), (1, 2, 3))
assert.eq(f(c=9, a=1, b=2), (1, 2, 9))

# *args and **kwargs unpack at the call site.
assert.eq(f(*[1, 2]), (1, 2, 3))
assert.eq(f(1, **{"b": 2, "c": 9}), (1, 2, 9))
assert.eq(f(*(1,), **dict(b=2)), (1, 2, 3))

# Argument binding errors.
assert.fails(lambda: f(1), "missing 1 argument \\(b\\)")
assert.fails(lambda: f(1, 2, 3, 4), "accepts at most 3 positional arguments \\(4 given\\)")
assert.fails(lambda: f(1, 2, z=3), 'unexpected keyword argument "z"')
assert.fails(lambda: f(1, a=1, b=2), "got multiple values for parameter \"a\"")

def g():
    pass

assert.fails(lambda: g(1), "accepts no arguments \\(1 given\\)")

# Only callable values may be called.
assert.fails(lambda: (1)(), "invalid call of non-function \\(int\\)")
assert.fails(lambda: "f"(), "invalid call of non-function \\(string\\)")

# Any expression yielding a callable may be called: element lookups,
# call results, and chained suffixes.
assert.eq({"f": len}["f"]("abc"), 3)
assert.eq([len][0]("ab"), 2)

def make():
    return lambda: 1

assert.eq(make()(), 1)
assert.eq(["abc"][0][0].upper(), "A")
