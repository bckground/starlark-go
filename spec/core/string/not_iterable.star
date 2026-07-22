# spec: spec.md#strings

# Strings support len, membership, and indexing, but are not
# iterable: every construct that iterates rejects a string operand.
assert.eq(len("abc"), 3)
assert.true("a" in "abc")
assert.eq("abc"[1], "b")

def for_string():
    for x in "abc":
        pass

assert.fails(for_string, "string value is not iterable")
assert.fails(lambda: [x for x in "abc"], "string value is not iterable")

def args(*args):
    return args

assert.fails(lambda: args(*"abc"), "must be iterable, not string")
assert.fails(lambda: list("abc"), "got string, want iterable")
assert.fails(lambda: tuple("abc"), "got string, want iterable")
assert.fails(lambda: enumerate("ab"), "got string, want iterable")
assert.fails(lambda: sorted("abc"), "got string, want iterable")
assert.fails(lambda: reversed("abc"), "got string, want iterable")
assert.fails(lambda: zip("ab", "cd"), "not iterable: string")
assert.fails(lambda: all("abc"), "got string, want iterable")
assert.fails(lambda: any("abc"), "got string, want iterable")
assert.fails(lambda: [].extend("bc"), "got string, want iterable")
assert.fails(lambda: ",".join("abc"), "got string, want iterable")
assert.fails(lambda: dict(["ab"]), "not iterable .*string")

# Iterate explicitly via elems() or codepoints() instead.
assert.eq([x for x in "abc".elems()], ["a", "b", "c"])
