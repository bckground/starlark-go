# spec: spec.md#integers

# int requires an argument; floats truncate toward zero (see
# float/conversions.star); bools convert to 0 or 1.
assert.fails(int, "missing argument")
assert.eq(int(3), 3)
assert.eq(int(3.1), 3)
assert.eq(int(True), 1)

# An explicit base is permitted only for strings.
assert.fails(lambda: int(3, base=10), "non-string with explicit base")
assert.fails(lambda: int(True, 10), "non-string with explicit base")

# Base 10 is the default: leading zeros are not octal, and prefixes
# are invalid.
assert.eq(int("123"), 123)
assert.eq(int("-123"), -123)
assert.eq(int("0123"), 123)
assert.eq(int("+0"), 0)
assert.eq(int("100000000000000000000"), 10000000000 * 10000000000)
assert.fails(lambda: int("0x12"), "invalid literal with base 10")
assert.fails(lambda: int("0b0"), "invalid literal with base 10")

# An explicit base from 2 to 36; a matching prefix is permitted and
# redundant.
assert.eq(int("11", base=9), 10)
assert.eq(int("10011", base=2), 19)
assert.eq(int("123", 8), 83)
assert.eq(int("0o123", 8), 83)
assert.eq(int("00123", 8), 83)  # redundant zeros permitted
assert.eq(int("123", 7), 66)    # 1*49 + 2*7 + 3
assert.eq(int("12", 16), 18)
assert.eq(int("0x12", 16), 18)
assert.eq(int("0b0101", 2), 5)

# A prefix that does not match the explicit base is reinterpreted as
# digits where valid, and an error otherwise.
assert.eq(int("0b0", 16), 0xB0)
assert.eq(int("0x0b0", 16), 0xB0)
assert.fails(lambda: int("0x123", 8), "invalid literal.*base 8")
assert.fails(lambda: int("0o123", 16), "invalid literal.*base 16")
assert.fails(lambda: int("0x110", 2), "invalid literal.*base 2")

# Base 0 detects the base from the prefix; redundant leading zeros
# are not permitted.
assert.eq(int("123", 0), 123)
assert.eq(int("-0x12", 0), -18)
assert.eq(int("0o123", 0), 83)
assert.eq(int("0b0101", 0), 5)
assert.fails(lambda: int("0123", 0), "invalid literal.*base 0")

# Malformed signs and digits.
assert.fails(lambda: int("--4"), "invalid literal with base 10: --4")
assert.fails(lambda: int("++4"), "invalid literal with base 10: \\+\\+4")
assert.fails(lambda: int("+-4"), "invalid literal with base 10: \\+-4")
assert.fails(lambda: int("0x-4", 16), "invalid literal with base 16: 0x-4")
