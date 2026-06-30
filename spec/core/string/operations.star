# spec: spec.md#strings

# A string is an immutable sequence of bytes holding (usually) UTF-8
# encoded text. len reports bytes, not code points.
assert.eq(type("abc"), "string")
assert.eq(len("abc"), 3)
assert.eq(len("é"), 2)

# Concatenation and repetition.
assert.eq("ab" + "cd", "abcd")
assert.eq("ab" * 3, "ababab")
assert.eq(3 * "ab", "ababab")
assert.eq("ab" * 0, "")

# Indexing yields a 1-byte substring; negative indices count from the
# end; out-of-range indexing is an error.
s = "hello"
assert.eq(s[0], "h")
assert.eq(s[-1], "o")
assert.eq(type(s[0]), "string")
assert.fails(lambda: s[5], "out of range")

# Slicing.
assert.eq(s[1:3], "el")
assert.eq(s[:2], "he")
assert.eq(s[2:], "llo")
assert.eq(s[:], "hello")
assert.eq(s[::2], "hlo")
assert.eq(s[::-1], "olleh")
assert.eq(s[1:100], "ello")  # slice indices are clamped

# x in s is a substring test.
assert.true("ell" in s)
assert.true("" in s)
assert.true("z" not in s)

# Strings are ordered lexicographically by bytes.
assert.lt("abc", "abd")
assert.lt("ab", "abc")
assert.lt("Z", "a")

# Strings are not iterable; use elems() or codepoints().
def iterate():
    return [c for c in "abc"]

assert.fails(iterate, "string value is not iterable")
assert.eq(list("abc".elems()), ["a", "b", "c"])
assert.eq(list("aé".codepoints()), ["a", "é"])
assert.eq(list("abc".elem_ords()), [97, 98, 99])
assert.eq(list("aé".codepoint_ords()), [97, 233])

# Repetition by a non-positive count yields the empty string.
assert.eq("abc" * -1, "")
assert.eq(-1 * "abc", "")
assert.fails(lambda: 1.0 * "abc", "unknown.*float \\* str")

# Strings do not support element assignment.
def set_char():
    "abc"[1] = "B"

assert.fails(set_char, "string.*does not support.*assignment")

# Non-unit strides clamp like slices.
assert.eq("banana"[7::-2], "aaa")
assert.eq("banana"[4::-2], "nnb")
assert.eq("banana"[None:None:-2], "aaa")
assert.fails(lambda: "banana"[1.0::], "invalid start index: got float, want int")
assert.fails(lambda: "banana"[:"":], "invalid end index: got string, want int")
assert.fails(lambda: "banana"[:"":True], "invalid slice step: got bool, want int")

# The in operator requires string operands.
assert.fails(lambda: 1 in "", "requires string as left operand")
assert.fails(lambda: "" in 1, "unknown binary op: string in int")

# An unknown method suggests the nearest name.
assert.fails(lambda: "".starts_with, "no .starts_with field.*did you mean .startswith")

# Strings are hashable; hash follows Java's String.hashCode.
d = {"k": 1}
assert.eq(d["k"], 1)
assert.eq(hash("\0" * 100), 0)
assert.eq(hash("hello"), 99162322)
assert.eq(hash("world"), 113318802)
assert.eq(hash("Hello, 世界!"), 417292677)
