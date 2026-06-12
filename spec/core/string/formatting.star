# spec: spec.md#strings

# The % operator formats a string. A tuple supplies multiple
# operands; any other value is a single operand.
assert.eq("%d" % 3, "3")
assert.eq("%s-%s" % ("a", "b"), "a-b")
assert.eq("%s" % "x", "x")
assert.eq("%r" % "x", '"x"')
assert.eq("%x" % 255, "ff")
assert.eq("%X" % 255, "FF")
assert.eq("%o" % 8, "10")
assert.eq("%%" % (), "%")

# Mismatched conversions are errors.
assert.fails(lambda: "%d" % "abc", "%d format requires integer")
assert.fails(lambda: "%d %d" % 1, "not enough arguments")

# The format method interpolates {} placeholders.
assert.eq("{} {}".format(1, 2), "1 2")
assert.eq("{1} {0}".format("a", "b"), "b a")
assert.eq("{x}/{y}".format(x=1, y=2), "1/2")
assert.eq("{{}}".format(), "{}")  # escaped braces
assert.eq("{!r}".format("a"), '"a"')

# Automatic and manual numbering may not be mixed.
assert.fails(lambda: "{}{1}".format(1, 2), "cannot switch from automatic field numbering to manual")

# A dict operand serves named conversions; unused keys are ignored.
assert.eq("A %(foo)d %(bar)s Z" % {"foo": 123, "bar": "hi"}, "A 123 hi Z")
assert.eq("A" % {"unused": 123}, "A")

# Operand count must match exactly.
assert.fails(lambda: "%d %d" % (1, 2, 3), "too many arguments for format string")
assert.fails(lambda: "" % 1, "too many arguments for format string")

# %c formats a code point, given as an int or a 1-code-point string.
assert.eq("%c" % 65, "A")
assert.eq("%c" % 0x3B1, "α")
assert.eq("%c" % "α", "α")
assert.fails(lambda: "%c" % "abc", "requires a single-character string")
assert.fails(lambda: "%c" % "", "requires a single-character string")
assert.fails(lambda: "%c" % 65.0, "requires int or single-character string")
assert.fails(lambda: "%c" % -1, "requires a valid Unicode code point")

# format: indices are decimal (leading zeros allowed), and the dotted
# or indexed field syntax is not supported.
assert.eq("{0000000000001}".format(0, 1), "1")
assert.eq("a{010}b".format(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10), "a10b")
assert.fails(lambda: "a{123}b".format(), "tuple index out of range")
assert.fails(lambda: "a{}b{}c".format(1), "tuple index out of range")
assert.fails(lambda: "a{z}b".format(x=1), "keyword z not found")
assert.fails(lambda: "{+1}".format(1), "keyword \\+1 not found")
assert.fails(lambda: "{a.b}".format(1), "syntax x.y is not supported")
assert.fails(lambda: "{a[0]}".format(1), "syntax a\\[i\\] is not supported")
assert.fails(lambda: "{ {} }".format(1), "nested replacement fields not supported")
assert.fails(lambda: "{x!}".format(x=1), "unknown conversion")

# Unbalanced braces are errors.
assert.fails(lambda: "{{}".format(1), "single '}' in format")
assert.fails(lambda: "{}}".format(1), "single '}' in format")
assert.fails(lambda: "}}{".format(1), "unmatched '{' in format")
