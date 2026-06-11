# Precise builtin result types flowing into errors.

xs = sorted([3, 1, 2])
s = xs[0] + "a"

z = zip([1], ["a"])
first: tuple[str, int] = z[0]

d = dict(a=1)
v = d.get("a", 2) + []

m = max(1, 2)
n = m + "a"
