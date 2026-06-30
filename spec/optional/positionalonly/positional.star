# spec: spec.md#positional-only-parameters

# Parameters before / are positional-only.
def f(x, y, /, z):
    return (x, y, z)

assert.eq(f(1, 2, 3), (1, 2, 3))
assert.eq(f(1, 2, z=3), (1, 2, 3))
assert.fails(lambda: f(1, y=2, z=3), 'got a value for positional-only parameter "y"')
assert.fails(lambda: f(x=1, y=2, z=3), 'got a value for positional-only parameter "x"')

# Positional-only parameters may have defaults.
def g(x, y=2, /):
    return (x, y)

assert.eq(g(1), (1, 2))
assert.eq(g(1, 9), (1, 9))

# / composes with *args, keyword-only parameters, and **kwargs.
def h(a, /, b, *, c, **kw):
    return (a, b, c, kw)

assert.eq(h(1, 2, c=3, d=4), (1, 2, 3, {"d": 4}))
assert.eq(h(1, b=2, c=3), (1, 2, 3, {}))
