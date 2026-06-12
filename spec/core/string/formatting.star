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
