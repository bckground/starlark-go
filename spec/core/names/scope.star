# spec: spec.md#name-binding-and-variables

# Within a function, a name may be used before the statement that
# binds it, as long as the binding executes first.
def forward():
    def helper():
        return later()

    def later():
        return 42

    return helper()

assert.eq(forward(), 42)

# Using a local before it is assigned is a dynamic error, even if a
# global of the same name exists.
g = 1

def use_before_assign():
    x = y
    y = 2

assert.fails(use_before_assign, "local variable y referenced before assignment")

def shadow_before_assign():
    x = g  # g is local here, because it is assigned below
    g = 2

assert.fails(shadow_before_assign, "local variable g referenced before assignment")

# An assignment anywhere in a function makes the name local
# throughout; reading the global is otherwise fine.
def read_global():
    return g

assert.eq(read_global(), 1)

# Assigning in a function does not affect the global.
def assign_local():
    h = 99
    return h

h = 1
assert.eq(assign_local(), 99)
assert.eq(h, 1)

# The variable of a comprehension is local to the comprehension.
squares = [i * i for i in range(3)]
assert.eq(squares, [0, 1, 4])
# (See resolve_errors.star: using i here would be a static error.)
