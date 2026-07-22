# spec: spec.md#booleans

# There are exactly two boolean values, True and False.
assert.eq(type(True), "bool")
assert.eq(type(False), "bool")
assert.ne(True, False)

# Booleans are ordered: False < True.
assert.lt(False, True)

# Booleans are not numbers: arithmetic on them is an error, but they
# convert explicitly.
assert.fails(lambda: True + 1, "unknown binary op: bool \\+ int")
assert.eq(int(True), 1)
assert.eq(int(False), 0)

# bool(x) reports the truth value of x.
assert.eq(bool(None), False)
assert.eq(bool(0), False)
assert.eq(bool(0.0), False)
assert.eq(bool(""), False)
assert.eq(bool([]), False)
assert.eq(bool(()), False)
assert.eq(bool({}), False)
assert.eq(bool(1), True)
assert.eq(bool(-1), True)
assert.eq(bool(0.1), True)
assert.eq(bool("False"), True)
assert.eq(bool([0]), True)
assert.eq(bool((None,)), True)
assert.eq(bool({0: 0}), True)

# not, and, or operate on truth values; and/or yield an operand, not
# a bool, and short-circuit.
assert.eq(not True, False)
assert.eq(not 0, True)
assert.eq(1 and 2, 2)
assert.eq(0 and 2, 0)
assert.eq(1 or 2, 1)
assert.eq(0 or 2, 2)

def boom():
    return 1 // 0

assert.eq(False and boom(), False)
assert.eq(True or boom(), True)

# Comparison operators yield booleans.
assert.eq(1 < 2, True)
assert.eq(1 == 2, False)
