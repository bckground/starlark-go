# spec: spec.md#bytes

# A bytes is an immutable array of integers in the range 0-255.
b = b"abc"
assert.eq(type(b), "bytes")
assert.eq(len(b), 3)

# Indexing yields the byte's value as an integer; slicing yields bytes.
assert.eq(b[0], 97)
assert.eq(b[-1], 99)
assert.eq(b[1:], b"bc")
assert.eq(type(b[1:2]), "bytes")

# Concatenation.
assert.eq(b"ab" + b"cd", b"abcd")

# x in b tests for an element when x is an int in [0, 255], and for a
# consecutive subsequence when x is bytes.
assert.true(97 in b)
assert.true(100 not in b)
assert.true(b"bc" in b)
assert.true(b"ac" not in b)

# Bytes are totally ordered and truthy when non-empty.
assert.lt(b"a", b"b")
assert.eq(bool(b""), False)
assert.eq(bool(b"\x00"), True)

# elems() yields the elements as ints.
assert.eq(list(b.elems()), [97, 98, 99])

# bytes(s) converts a string to its UTF-8 encoding.
assert.eq(bytes("A"), b"A")

# Bytes are hashable and usable as dict keys.
d = {b"k": 1}
assert.eq(d[b"k"], 1)
