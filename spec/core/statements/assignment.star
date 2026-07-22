# spec: spec.md#assignments

# An assignment may target a name, a list or tuple of targets, an
# element, or a field.
a = 1
assert.eq(a, 1)

b, c = 1, 2
assert.eq((b, c), (1, 2))

[d, e] = [3, 4]
assert.eq((d, e), (3, 4))

(f, (g, h)) = (1, (2, 3))
assert.eq((f, g, h), (1, 2, 3))

# List and tuple targets are interchangeable, and nest either way.
[i, j, (k, l)] = (1, 2, [3, 4])
assert.eq((i, j, k, l), (1, 2, 3, 4))
(m, [n, o]) = [1, (2, 3)]
assert.eq((m, n, o), (1, 2, 3))

# A bare tuple on the right matches a list target.
[p, q] = 1, 2
assert.eq((p, q), (1, 2))

# Singleton and parenthesized targets.
(single,) = [1]
assert.eq(single, 1)
(paren) = 1
assert.eq(paren, 1)

# The number of elements must match.
def mismatch():
    x, y = 1, 2, 3

assert.fails(mismatch, "too many values to unpack")

def mismatch2():
    x, y, z = 1, 2

assert.fails(mismatch2, "too few values to unpack")

# Element assignment.
def element_assignment():
    l = [1, 2]
    l[0] = 10
    m = {"k": 1}
    m["k"] = 2
    return (l, m)

assert.eq(element_assignment(), ([10, 2], {"k": 2}))

# Augmented assignment: x op= y is like x = x op y, but the target is
# evaluated once.
def augmented():
    n = 1
    n += 2
    n *= 3
    n -= 1
    n //= 2
    return n

assert.eq(augmented(), 4)

# For a list, += mutates the list in place (aliases observe it);
# for an immutable sequence it rebinds the variable.
def list_in_place():
    l = [1]
    alias = l
    l += [2]
    return (l, alias)

assert.eq(list_in_place(), ([1, 2], [1, 2]))

def tuple_rebind():
    t = (1,)
    alias = t
    t += (2,)
    return (t, alias)

assert.eq(tuple_rebind(), ((1, 2), (1,)))

def element_augmented():
    counts = {"a": 0}
    counts["a"] += 1
    return counts

assert.eq(element_augmented(), {"a": 1})
