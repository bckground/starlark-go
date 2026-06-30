# spec: spec.md#integers

# Floored division by zero aborts execution.
x = 1 // 0 ### "floored division by zero"
---
# Modulo by zero aborts execution.
x = 1 % 0 ### "integer modulo by zero"
---
# An abort inside a function unwinds the whole stack.
def f():
    return 1 // 0 ### "floored division by zero"

f()
---
# When the aborting computation can be wrapped in a callable, the
# abort can be observed with assert.fails or trap, and the program
# itself completes without error.
assert.fails(lambda: 1 // 0, "floored division by zero")

msg = trap(lambda: 1 % 0)
assert.true(msg != None, "expected an abort")
assert.true(matches("integer modulo by zero", msg))
