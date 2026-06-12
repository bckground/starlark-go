# spec: spec.md#reassigning-globals

# With both units, top-level loops may rebind globals, enabling
# accumulation without a function.
total = 0
for x in [1, 2, 3]:
    total += x

assert.eq(total, 6)

# Both branches of a top-level if may bind the same name.
if total > 5:
    size = "big"
else:
    size = "small"

assert.eq(size, "big")
