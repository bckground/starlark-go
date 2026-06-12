# spec: spec.md#function-definitions

# A def statement binds a function value.
def add(x, y):
    return x + y

assert.eq(type(add), "function")
assert.eq(str(add), "<function add>")
assert.eq(add(1, 2), 3)

# Parameters may have default values, evaluated once, when the def
# statement is executed.
def inc(x, by=1):
    return x + by

assert.eq(inc(3), 4)
assert.eq(inc(3, 10), 13)

# A mutable default value is shared between calls.
def remember(x, seen=[]):
    seen.append(x)
    return seen

assert.eq(remember(1), [1])
assert.eq(remember(2), [1, 2])

# *args collects surplus positional arguments as a tuple; **kwargs
# collects surplus keyword arguments as a dict; parameters after *
# are keyword-only.
def variadic(a, *args, key=0, **kwargs):
    return (a, args, key, kwargs)

assert.eq(variadic(1), (1, (), 0, {}))
assert.eq(variadic(1, 2, 3, key=4, extra=5), (1, (2, 3), 4, {"extra": 5}))

def kwonly(a, *, k):
    return (a, k)

assert.eq(kwonly(1, k=2), (1, 2))
assert.fails(lambda: kwonly(1, 2), "accepts 1 positional argument")

# Nested functions close over enclosing local variables.
def outer(n):
    def inner(m):
        return n + m

    return inner

add10 = outer(10)
assert.eq(add10(5), 15)

# Recursion is an error.
def recurse():
    return recurse()

assert.fails(recurse, "function recurse called recursively")

def even(n):
    return True if n == 0 else odd(n - 1)

def odd(n):
    return False if n == 0 else even(n - 1)

assert.fails(lambda: even(4), "function even called recursively")
