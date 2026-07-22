# Tests of positional-only parameters: def f(x, /, y).
# option:positionalonly

load("assert.star", "assert")

def f(x, /, y):
    return (x, y)

assert.eq(f(1, 2), (1, 2))
assert.eq(f(1, y=2), (1, 2))
assert.fails(lambda: f(x=1, y=2), 'got a value for positional-only parameter "x"')
assert.fails(lambda: f(1, 2, 3), "accepts 2 positional arguments")

# defaults work on both sides of the marker
def g(x, y=1, /, z=2):
    return (x, y, z)

assert.eq(g(0), (0, 1, 2))
assert.eq(g(0, 5), (0, 5, 2))
assert.eq(g(0, 5, z=9), (0, 5, 9))
assert.fails(lambda: g(0, y=5), 'got a value for positional-only parameter "y"')

# like Python, a positional-only parameter's name remains available
# to **kwargs
def h(x, /, **kwargs):
    return (x, kwargs)

assert.eq(h(1, x=2), (1, {"x": 2}))
assert.eq(h(1, y=3), (1, {"y": 3}))

# combined with * and keyword-only parameters
def combo(a, b, /, c, *, d):
    return (a, b, c, d)

assert.eq(combo(1, 2, 3, d=4), (1, 2, 3, 4))
assert.eq(combo(1, 2, c=3, d=4), (1, 2, 3, 4))
assert.fails(lambda: combo(1, b=2, c=3, d=4), 'got a value for positional-only parameter "b"')

# lambdas support the marker too
pick = lambda x, /, y: y
assert.eq(pick(1, y=2), 2)
assert.fails(lambda: pick(x=1, y=2), 'got a value for positional-only parameter "x"')

---
# positional-only markers combine with type annotations
# option:positionalonly option:types

load("assert.star", "assert")

def scale(x: int, /, factor: int = 2) -> int:
    return x * factor

assert.eq(scale(3), 6)
assert.eq(scale(3, factor=3), 9)
assert.fails(lambda: scale("a"), "does not match the type annotation `int` for argument `x`")
assert.fails(lambda: scale(x=3), 'got a value for positional-only parameter "x"')
