# spec: spec.md#floating-point-numbers

nan = float("NaN")
inf = float("+Inf")
neginf = float("-Inf")
negzero = -(0.0)

# Arithmetic with infinities.
assert.eq(inf + 1, inf)
assert.eq(inf + inf, inf)
assert.eq(inf - 1, inf)
assert.eq(inf - neginf, inf)
assert.eq(inf * 1, inf)
assert.eq(inf * inf, inf)
assert.eq(inf * neginf, neginf)
assert.eq(-(inf), neginf)
assert.eq(-neginf, inf)
assert.eq(0.0 / inf, 0.0)
assert.eq(0.0 / neginf, 0.0)

# Indeterminate forms yield NaN.
assert.eq(inf + neginf, nan)
assert.eq(inf - inf, nan)
assert.eq(inf * 0, nan)
assert.eq(inf / inf, nan)
assert.eq(inf / neginf, nan)
assert.eq(inf % 1, nan)
assert.eq(inf % inf, nan)

# NaN propagates through arithmetic.
assert.eq(1.2e3 + nan, nan)
assert.eq(1.2e3 - nan, nan)
assert.eq(1.2e3 * nan, nan)
assert.eq(100.0 / nan, nan)
assert.eq(100.0 // nan, nan)
assert.eq(1.2e3 % nan, nan)
assert.eq(str(-(nan)), "nan")

# The total order: NaN equals itself and is the greatest float;
# unlike IEEE 754, == and != never both fail.
assert.true(nan == nan)
assert.true(nan >= nan)
assert.true(nan <= nan)
assert.true(not (nan > nan))
assert.true(not (nan < nan))
assert.lt(inf, nan)
assert.lt(1.7976931348623157e+308, inf)  # approx. largest finite float
assert.lt(neginf, -1.7976931348623157e+308)
assert.lt(4.9406564584124654e-324, 1.0)  # approx. smallest positive float
assert.lt(0.0, 4.9406564584124654e-324)

# Negative zero equals positive zero but prints differently.
assert.eq(negzero, 0.0)
assert.eq(negzero, 0)
assert.eq(str(negzero), "-0.0")
assert.eq(str(1 // neginf), "-0.0")

# max and min follow the total order: NaN is the maximum.
assert.eq(max([1, nan, 3]), nan)
assert.eq(min([nan, 2, 3]), 2)
assert.eq(min([1, nan, 3]), 1)

# Sorting follows the total order and is stable: equal values such as
# 0.0 and -0.0, or 1 and 1.0, keep their relative order.
assert.eq(str(sorted([inf, neginf, nan, 1.0, -1.0, 1, -1, 0, 0.0, negzero])),
          "[-inf, -1.0, -1, 0, 0.0, -0.0, 1.0, 1, +inf, nan]")
assert.eq(str(sorted([7, 3, nan, 1, 9])), "[1, 3, 7, 9, nan]")
assert.eq(str(sorted([7, 3, nan, 1, 9], reverse=True)), "[nan, 9, 7, 3, 1]")

# All NaN values are the same dict key, even from distinct
# evaluations.
nandict = {nan: 1}
nandict[nan] = 2
assert.eq(len(nandict), 1)
assert.eq(nandict[nan], 2)
nandict[float("nan")] = 3
assert.eq(len(nandict), 1)
assert.fails(lambda: {nan: 1, nan: 2}, "duplicate key: nan")

# Infinities are ordinary dict keys.
assert.eq(str({inf: 1, neginf: 2}), "{+inf: 1, -inf: 2}")

# NaN is true; zero is false regardless of sign.
assert.true(nan)
assert.true(not negzero)
