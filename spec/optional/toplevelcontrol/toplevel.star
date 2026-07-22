# spec: spec.md#top-level-control-flow

# if statements may appear at module level. (Without the
# globalreassign unit, each global may still be bound by only one
# statement, so the two branches must bind different names.)
if 1 < 2:
    branch = "then"

assert.eq(branch, "then")

# for statements may appear at module level. A loop binds its
# variable by a single statement, so iteration is permitted; mutation
# accumulates results without rebinding.
out = []
for x in [1, 2, 3]:
    out.append(x * x)

assert.eq(out, [1, 4, 9])
assert.eq(x, 3)  # the loop variable remains bound after the loop

# Control statements nest at module level.
nested = []
for i in range(3):
    if i != 1:
        nested.append(i)

assert.eq(nested, [0, 2])
