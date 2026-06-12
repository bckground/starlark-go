# spec: spec.md#floating-point-numbers

nan = float("NaN")
inf = float("+Inf")
neginf = float("-Inf")
negzero = -(0.0)

# str of a float always shows a decimal point or an exponent, so the
# value reads back as a float.
assert.eq(str(0.0), "0.0")
assert.eq(str(0.), "0.0")
assert.eq(str(.0), "0.0")
assert.eq(str(123.0), "123.0")
assert.eq(str(1.23e45), "1.23e+45")
assert.eq(str(-1.23e-45), "-1.23e-45")
assert.eq(str(nan), "nan")
assert.eq(str(inf), "+inf")
assert.eq(str(neginf), "-inf")
assert.eq(str(negzero), "-0.0")

# %d accepts an integral float and converts it exactly; non-finite
# values are errors.
assert.eq("%d" % 0.0, "0")
assert.eq("%d" % 123.0, "123")
assert.eq("%d" % negzero, "0")
assert.eq("%d" % 1.23e45, "1229999999999999973814869011019624571608236032")
assert.fails(lambda: "%d" % nan, "cannot convert float NaN to integer")
assert.fails(lambda: "%d" % inf, "cannot convert float infinity to integer")

# %e: scientific notation with six fractional digits.
assert.eq("%e" % 0.0, "0.000000e+00")
assert.eq("%e" % 123, "1.230000e+02")
assert.eq("%e" % 1.23e45, "1.230000e+45")
assert.eq("%e" % -1.23e-45, "-1.230000e-45")
assert.eq("%e" % nan, "nan")
assert.eq("%e" % inf, "+inf")
assert.eq("%e" % negzero, "-0.000000e+00")
assert.fails(lambda: "%e" % "123", "requires float, not str")

# %f: fixed-point with six fractional digits, however large.
assert.eq("%f" % 0.0, "0.000000")
assert.eq("%f" % 123, "123.000000")
assert.eq("%f" % 1.23e45, "1229999999999999973814869011019624571608236032.000000")
assert.eq("%f" % -1.23e-45, "-0.000000")
assert.eq("%f" % nan, "nan")
assert.eq("%f" % neginf, "-inf")

# %g: shortest form that round-trips; large magnitudes switch to
# scientific notation. str uses %g.
assert.eq("%g" % 0.0, "0.0")
assert.eq("%g" % 123.0, "123.0")
assert.eq("%g" % 1.110, "1.11")
assert.eq("%g" % 1e5, "100000.0")
assert.eq("%g" % 1e6, "1e+06")
assert.eq("%g" % 1.23e45, "1.23e+45")
assert.eq("%g" % nan, "nan")
assert.eq("%g" % inf, "+inf")
assert.eq("%g" % negzero, "-0.0")

# %e/%f/%g accept ints by conversion, but not strings.
assert.eq("%e" % 0, "0.000000e+00")
assert.eq("%g" % 123, "123.0")
assert.fails(lambda: "%g" % "123", "requires float, not str")
