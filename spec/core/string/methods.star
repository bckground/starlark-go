# spec: spec.md#strings

# Case conversion.
assert.eq("aBc".upper(), "ABC")
assert.eq("aBc".lower(), "abc")
assert.eq("hello world".capitalize(), "Hello world")
assert.eq("hello world".title(), "Hello World")

# Predicates.
assert.true("abc123".isalnum())
assert.true(not "abc!".isalnum())
assert.true("abc".isalpha())
assert.true("123".isdigit())
assert.true("abc".islower())
assert.true("ABC".isupper())
assert.true(" \t\n".isspace())
assert.true("Hello World".istitle())
assert.true("hello".startswith("he"))
assert.true("hello".endswith("lo"))
assert.true(not "hello".startswith("lo"))

# Search.
assert.eq("hello".find("l"), 2)
assert.eq("hello".find("z"), -1)
assert.eq("hello".rfind("l"), 3)
assert.eq("hello".index("l"), 2)
assert.fails(lambda: "hello".index("z"), "substring not found")
assert.eq("hello".count("l"), 2)
assert.eq("aaa".count("aa"), 1)  # non-overlapping

# Replace.
assert.eq("banana".replace("a", "o"), "bonono")
assert.eq("banana".replace("a", "o", 2), "bonona")

# Split and join.
assert.eq("a,b,c".split(","), ["a", "b", "c"])
assert.eq("a,b,c".split(",", 1), ["a", "b,c"])
assert.eq("a,b,c".rsplit(",", 1), ["a,b", "c"])
assert.eq("  a  b  ".split(), ["a", "b"])  # default: runs of whitespace
assert.eq("-".join(["a", "b", "c"]), "a-b-c")
assert.eq("".join([]), "")
assert.eq("a\nb\nc".splitlines(), ["a", "b", "c"])
assert.eq("a\nb\n".splitlines(True), ["a\n", "b\n"])

# Strip.
assert.eq("  hi  ".strip(), "hi")
assert.eq("  hi  ".lstrip(), "hi  ")
assert.eq("  hi  ".rstrip(), "  hi")
assert.eq("xxhixx".strip("x"), "hi")

# Partition.
assert.eq("a=b=c".partition("="), ("a", "=", "b=c"))
assert.eq("a=b=c".rpartition("="), ("a=b", "=", "c"))
assert.eq("abc".partition("="), ("abc", "", ""))
