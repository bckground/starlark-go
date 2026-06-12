# spec: spec.md#functions

# Each call of a function creates fresh local bindings: closures made
# by different calls are independent.
def outer(x):
    def inner(y):
        return x + y

    return inner

add6 = outer(6)
add8 = outer(8)
assert.eq(add6(5), 11)
assert.eq(add8(5), 13)
assert.eq(add6(7), 13)  # add6 is unaffected by the later call

# A closure may carry mutable state.
def counter():
    state = [0]

    def next():
        state[0] += 1
        return state[0] * state[0]

    return next

sq = counter()
assert.eq(sq(), 1)
assert.eq(sq(), 4)
assert.eq(sq(), 9)

# Freezing a function freezes the values it closes over...
frozen_sq = freeze(counter())
assert.fails(frozen_sq, "frozen list")

# ...and its parameter defaults. An unfrozen mutable default is
# shared and mutable across calls.
def dflt(x=[0]):
    return x

dflt().append(1)
assert.eq(dflt(), [0, 1])
freeze(dflt)
assert.fails(lambda: dflt().append(2), "cannot append to frozen list")

# str of a function names it.
assert.eq(str(outer), "<function outer>")
assert.eq(str(add6), "<function inner>")
