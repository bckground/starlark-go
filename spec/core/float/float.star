# spec: spec.md#floating-point-numbers

# Floating-point numbers use IEEE 754 double precision.
assert.eq(type(1.0), "float")
assert.eq(type(1e3), "float")
assert.eq(1e3, 1000.0)
assert.eq(0.25 + 0.25, 0.5)
assert.eq(1.5 * 2, 3.0)

# Mixed int/float arithmetic yields float.
assert.eq(type(1 + 0.5), "float")
assert.eq(1 + 0.5, 1.5)

# An int and a float compare equal if they denote the same value.
assert.eq(1.0, 1)
assert.eq(0.0, 0)
assert.lt(1, 1.5)
assert.lt(1.5, 2)

# Division.
assert.eq(7.0 / 2, 3.5)
assert.eq(7.0 // 2, 3.0)
assert.eq(type(7.0 // 2), "float")
assert.eq(-7.0 // 2, -4.0)
assert.eq(7.0 % 2, 1.0)
assert.fails(lambda: 1.0 / 0, "floating-point division by zero")
assert.fails(lambda: 1 / 0.0, "floating-point division by zero")
assert.fails(lambda: 1.0 // 0, "floored division by zero")
assert.fails(lambda: 1.0 % 0, "floating-point modulo by zero")

# float(x) converts a number or a decimal string.
assert.eq(float(3), 3.0)
assert.eq(float("1.5"), 1.5)
assert.eq(float(True), 1.0)
assert.fails(lambda: float("pi"), "invalid float literal")

# Unlike IEEE 754, comparison defines a total order on floats:
# NaN compares equal to itself and greater than +infinity.
nan = float("nan")
inf = float("inf")
assert.eq(nan, nan)
assert.lt(inf, nan)
assert.lt(1e308, inf)
assert.lt(-inf, 0)

# Floored division and remainder follow the sign rules of their int
# counterparts: // rounds toward negative infinity, % takes the sign
# of the divisor.
assert.eq(100.0 // 8.0, 12.0)
assert.eq(100.0 // -8.0, -13.0)
assert.eq(-100.0 // 8.0, -13.0)
assert.eq(-100.0 // -8.0, 12.0)
assert.eq(100.0 % 8.0, 4.0)
assert.eq(100.0 % -8.0, -4.0)
assert.eq(-100.0 % 8.0, 4.0)
assert.eq(-100.0 % -8.0, -4.0)
assert.eq(98.0 % -8.0, -6.0)
assert.eq(-98.0 % 8.0, 6.0)

# If either operand is a float the result is a float; / yields a
# float even for two ints.
ops = {
    "+": lambda x, y: x + y,
    "-": lambda x, y: x - y,
    "*": lambda x, y: x * y,
    "/": lambda x, y: x / y,
    "//": lambda x, y: x // y,
    "%": lambda x, y: x % y,
}

def check_result_types():
    for name in ops:
        for x in (1, 1.0):
            for y in (1, 1.0):
                if name == "/" or type(x) == "float" or type(y) == "float":
                    want = "float"
                else:
                    want = "int"
                got = type(ops[name](x, y))
                assert.true(got == want, "%s %s %s yields %s, want %s" % (type(x), name, type(y), got, want))

check_result_types()
