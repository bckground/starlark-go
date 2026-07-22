# spec: spec.md#lexical-elements

# Decimal, hexadecimal, octal, and binary integer literals.
assert.eq(255, 0xFF)
assert.eq(255, 0xff)
assert.eq(255, 0o377)
assert.eq(255, 0b11111111)
assert.eq(0, 0x0)

# Literals may be arbitrarily large.
assert.eq(123456789012345678901234567890 + 0, 123456789012345678901234567890)

# Float literals.
assert.eq(1., 1.0)
assert.eq(.5, 0.5)
assert.eq(1e2, 100.0)
assert.eq(1.5e2, 150.0)
assert.eq(1e-2, 0.01)
