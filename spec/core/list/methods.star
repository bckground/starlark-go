# spec: spec.md#lists

# append adds one element at the end and returns None.
l = [1, 2]
assert.eq(l.append(3), None)
assert.eq(l, [1, 2, 3])

# extend appends each element of an iterable.
l2 = [1]
l2.extend([2, 3])
l2.extend((4,))
assert.eq(l2, [1, 2, 3, 4])

# insert inserts at a position; an out-of-range position clamps.
l3 = [1, 3]
l3.insert(1, 2)
assert.eq(l3, [1, 2, 3])
l3.insert(100, 4)
assert.eq(l3, [1, 2, 3, 4])
l3.insert(-100, 0)
assert.eq(l3, [0, 1, 2, 3, 4])

# remove removes the first occurrence; absent values are an error.
l4 = [1, 2, 1]
l4.remove(1)
assert.eq(l4, [2, 1])
assert.fails(lambda: l4.remove(9), "remove: element not found")

# pop removes and returns the element at an index (default: last).
l5 = [1, 2, 3]
assert.eq(l5.pop(), 3)
assert.eq(l5.pop(0), 1)
assert.eq(l5, [2])
assert.fails(lambda: [].pop(), "index -1 out of range: empty list")

# index finds the first occurrence; absent values are an error.
l6 = [1, 2, 1]
assert.eq(l6.index(1), 0)
assert.eq(l6.index(1, 1), 2)  # optional start
assert.fails(lambda: l6.index(9), "value not in list")

# clear removes all elements.
l7 = [1, 2]
l7.clear()
assert.eq(l7, [])
