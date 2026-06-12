# spec: spec.md#comprehensions

# List comprehensions.
assert.eq([x * x for x in range(4)], [0, 1, 4, 9])
assert.eq([x for x in [1, 2, 3, 4] if x % 2 == 0], [2, 4])

# Multiple for clauses and filters compose left to right.
assert.eq([(x, y) for x in range(2) for y in range(2)],
          [(0, 0), (0, 1), (1, 0), (1, 1)])
assert.eq([x + y for x in [1, 2] for y in [10] if x > 1], [12])

# The iteration variable may destructure.
assert.eq([k + str(v) for k, v in [("a", 1), ("b", 2)]], ["a1", "b2"])

# Dict comprehensions.
assert.eq({x: x * x for x in range(3)}, {0: 0, 1: 1, 2: 4})
assert.eq({v: k for k, v in {"a": 1}.items()}, {1: "a"})

# Comprehensions nest.
assert.eq([[y for y in range(x)] for x in range(3)], [[], [0], [0, 1]])

# The iteration variable is local to the comprehension and shadows
# enclosing bindings.
x = "outer"
assert.eq([x for x in [1]], [1])
assert.eq(x, "outer")
