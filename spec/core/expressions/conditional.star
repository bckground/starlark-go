# spec: spec.md#conditional-expressions

# a if cond else b evaluates only the chosen branch.
assert.eq(1 if True else 2, 1)
assert.eq(1 if False else 2, 2)
assert.eq("yes" if [0] else "no", "yes")  # condition uses truth value

def boom():
    return 1 // 0

assert.eq(1 if True else boom(), 1)
assert.eq(2 if False else 2, 2)

# Conditional expressions nest.
def sign(x):
    return -1 if x < 0 else (0 if x == 0 else 1)

assert.eq(sign(-5), -1)
assert.eq(sign(0), 0)
assert.eq(sign(5), 1)
