# spec: spec.md#built-in-constants-and-functions

# int converts a number or string. Conversion of a float truncates
# toward zero.
assert.eq(int(3.9), 3)
assert.eq(int(-3.9), -3)
assert.eq(int("12"), 12)
assert.eq(int("-12"), -12)
assert.eq(int("ff", 16), 255)
assert.eq(int("0x10", 16), 16)
assert.eq(int("101", 2), 5)
assert.eq(int(True), 1)
assert.fails(lambda: int("abc"), "invalid literal")
assert.fails(lambda: int(float("nan")), "cannot convert")

# str yields the string form of a value; repr quotes strings inside.
assert.eq(str(42), "42")
assert.eq(str("x"), "x")
assert.eq(str([1, "a"]), '[1, "a"]')
assert.eq(str((1,)), "(1,)")
assert.eq(str(None), "None")
assert.eq(str(True), "True")
assert.eq(repr("x"), '"x"')
assert.eq(repr([1, "a"]), '[1, "a"]')

# type yields the name of a value's type as a string.
assert.eq(type(None), "NoneType")
assert.eq(type(True), "bool")
assert.eq(type(1), "int")
assert.eq(type(1.0), "float")
assert.eq(type(""), "string")
assert.eq(type(b""), "bytes")
assert.eq(type([]), "list")
assert.eq(type(()), "tuple")
assert.eq(type({}), "dict")
assert.eq(type(range(1)), "range")
assert.eq(type(len), "builtin_function_or_method")

# chr and ord convert between code points and 1-code-point strings.
assert.eq(chr(65), "A")
assert.eq(ord("A"), 65)
assert.eq(ord(chr(0x1F600)), 0x1F600)
assert.fails(lambda: ord("ab"), "string encodes 2 Unicode code points, want 1")

# hash returns a deterministic 32-bit hash of a string, using the
# same algorithm as Java's String.hashCode.
assert.eq(hash(""), 0)
assert.eq(hash("abc"), 96354)
assert.fails(lambda: hash([1]), "")

# abs yields the magnitude of a number.
assert.eq(abs(-5), 5)
assert.eq(abs(5), 5)
assert.eq(abs(-1.5), 1.5)
assert.eq(abs(-1 << 100), 1 << 100)
