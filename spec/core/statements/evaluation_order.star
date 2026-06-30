# spec: spec.md#assignments

calls = []

def f(name, result):
    calls.append(name)
    return result

# In an ordinary assignment, the right side is evaluated before the
# left side's index and object expressions.
f("array", [0])[f("index", 0)] = f("rhs", 0)
assert.eq(calls, ["rhs", "array", "index"])

calls.clear()
f("lhs1", [0])[0], f("lhs2", [0])[0] = f("rhs1", 0), f("rhs2", 0)
assert.eq(calls, ["rhs1", "rhs2", "lhs1", "lhs2"])

# In an augmented assignment, the left side is evaluated first, and
# only once.
calls.clear()
f("array", [0])[f("index", 0)] += f("addend", 1)
assert.eq(calls, ["array", "index", "addend"])

# The single evaluation of the left side is observable.
count = [0]

def next():
    count[0] += 1
    return count[0]

x = [1, 2, 3]
x[next()] += 1
assert.eq(x, [1, 3, 3])  # the sole call to next returned 1
assert.eq(count[0], 1)
