# spec: spec.md#reassigning-globals

# A module-level name may be reassigned.
x = 1
x = 2
assert.eq(x, 2)

# Augmented assignment of a global works at module level.
x += 10
assert.eq(x, 12)

# Functions observe the global's value at call time.
def read():
    return y

y = "first"
assert.eq(read(), "first")
y = "second"
assert.eq(read(), "second")

# Function definitions may be replaced.
def f():
    return "old"

def f():
    return "new"

assert.eq(f(), "new")
