# spec: spec.md#comprehensions

# A comprehension is a single new lexical block, not one per clause:
# every 'in' operand except the first is resolved inside the block,
# so this x is the comprehension's own (unassigned) variable.
x = [1, 2]
_ = [x for _ in [3] for x in x] ### "local variable x referenced before assignment"
---
# The first 'in' operand, by contrast, is resolved in the enclosing
# block; this chunk must execute without error.
x = [[1, 2]]
assert.eq([x for x in x for y in x], [[1, 2], [1, 2]])
