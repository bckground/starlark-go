# Tests of the enum() builtin.

load("assert.star", "assert")

Color = enum("red", "green", "blue")

assert.eq(type(Color), "enum_type")
assert.eq(str(Color), 'enum("red", "green", "blue")')
assert.eq(len(Color), 3)
assert.eq(Color.values(), ["red", "green", "blue"])

c = Color("red")
assert.eq(type(c), "enum")
assert.eq(c.value, "red")
assert.eq(c.index, 0)
assert.eq(str(c), 'Color("red")')

# elements are interned; calling with an element returns it unchanged
assert.true(Color("red") == c)
assert.eq(Color(c), c)

# indexing and iteration
assert.eq(Color[1].value, "green")
assert.eq(Color[-1], Color("blue"))
assert.eq([x.value for x in Color], ["red", "green", "blue"])

# attribute access by element name
assert.eq(Color.red, Color("red"))
assert.eq(Color.blue.index, 2)

# equality is per-type
Shade = enum("red", "dark")
assert.ne(Color("red"), Shade("red"))
assert.eq(Shade("red").index, 0)

# enum values hash like their strings
d = {Color("red"): 1}
assert.eq(d[Color("red")], 1)

# errors
assert.fails(lambda: Color("purple"), 'unknown enum element "purple"')
assert.fails(lambda: Color(1), "got int, want string")
assert.fails(lambda: enum("a", "a"), 'duplicate value "a"')
assert.fails(lambda: enum(1), "got int, want string")
assert.fails(lambda: enum(), "at least one value is required")

# enum types are usable as type annotations
def paint(c: Color) -> str:
    return c.value

assert.eq(paint(Color("green")), "green")
assert.fails(lambda: paint("green"), "does not match the type annotation `Color`")
assert.fails(lambda: paint(Shade("red")), "does not match the type annotation `Color`")

# isinstance integration
assert.true(isinstance(c, Color))
assert.true(not isinstance("red", Color))
assert.true(isinstance(Color, type))

# assigning to a global names the type (export is by first assignment)
def make_enum():
    return enum("x", "y")

Named = make_enum()
assert.eq(str(Named("x")), 'Named("x")')

# a type that is never assigned to a global stays anonymous
def check_anon():
    anon = enum("x", "y")  # local: no export
    assert.eq(str(anon("x")), 'enum()("x")')

check_anon()
