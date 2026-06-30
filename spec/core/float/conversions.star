# spec: spec.md#floating-point-numbers

# float() parses decimal and scientific notation, with optional sign,
# and the case-insensitive special spellings.
assert.eq(float(), 0.0)
assert.eq(float("1.1"), 1.1)
assert.eq(float("-1.1"), -1.1)
assert.eq(float("+1.1"), 1.1)
assert.eq(float("+Inf"), float("INFINITY"))
assert.eq(str(float("Inf")), "+inf")
assert.eq(str(float("-iNfInItY")), "-inf")
assert.eq(str(float("+NAN")), "nan")
assert.eq(str(float("-nan")), "nan")
assert.fails(lambda: float("1.1abc"), "invalid float literal")
assert.fails(lambda: float("1e100.0"), "invalid float literal")
assert.fails(lambda: float("1.2.3"), "invalid float literal")
assert.fails(lambda: float("1e1000"), "floating-point number too large")
assert.fails(lambda: float(None), "want number or string")

# float of bool and int.
assert.eq(float(False), 0.0)
assert.eq(float(True), 1.0)
assert.eq(float(123), 123.0)
# But bools do not equal their float conversions.
assert.ne(False, 0.0)
assert.ne(True, 1.0)

# Conversion of an int rounds to the nearest representable float
# (round half to even); comparison, by contrast, is always exact.
p53 = 1 << 53
assert.eq(float(p53 - 1), p53 - 1)
assert.eq(float(p53 + 1), p53)
assert.eq(float(p53 + 3), p53 + 4)
assert.eq(float(p53 + 5), p53 + 4)
assert.eq(float(p53 + 7), p53 + 8)
assert.true(float(p53 + 1) != p53 + 1)   # comparison is exact
assert.eq(float(p53 + 1) - (p53 + 1), 0)  # arithmetic rounds first

# A float literal equals exactly one integer value.
assert.true(1.23e45 != 1229999999999999973814869011019624571608236031)
assert.true(1.23e45 == 1229999999999999973814869011019624571608236032)
assert.true(1.23e45 != 1229999999999999973814869011019624571608236033)

# An int too large for a float cannot convert, explicitly or
# implicitly.
huge = 1 << 500 << 500 << 500
assert.fails(lambda: float(huge), "int too large to convert to float")
assert.fails(lambda: huge + 0.0, "int too large to convert to float")
assert.fails(lambda: 0.0 * huge, "int too large to convert to float")
assert.fails(lambda: 1.0 // huge, "int too large to convert to float")
assert.fails(lambda: float(str(huge)), "floating-point number too large")

# int() of a float truncates toward zero; non-finite values cannot
# convert.
assert.eq(int(100.1), 100)
assert.eq(int(99.9), 99)
assert.eq(int(-99.9), -99)
assert.eq(int(-100.1), -100)
assert.eq(int(0.9), 0)
assert.eq(int(-0.9), 0)
assert.eq(int(1.23e-32), 0)
assert.eq(int(1.23e+32), 123000000000000004979083645550592)
assert.eq(int(1e100), int("10000000000000000159028911097599180468360808563945281389781327557747838772170381060813469985856815104"))
assert.fails(lambda: int(float("+Inf")), "cannot convert float infinity to integer")
assert.fails(lambda: int(float("-Inf")), "cannot convert float infinity to integer")
assert.fails(lambda: int(float("NaN")), "cannot convert float NaN to integer")

# Equal int and float values are the same dict key (same hash).
d = {123.0: "x"}
d[123] = "y"
assert.eq(len(d), 1)
assert.eq(d[123.0], "y")
assert.fails(lambda: {123.0: "f", 123: "i"}, "duplicate key: 123")

# Floats cannot be used as indices, even when integral.
assert.fails(lambda: "abc"[1.0], "want int")
assert.fails(lambda: ["A", "B", "C"].insert(1.0, "D"), "want int")
assert.fails(lambda: range(3)[1.0], "got float, want int")
