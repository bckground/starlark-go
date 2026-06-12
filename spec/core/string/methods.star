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
assert.eq("abc".rpartition("="), ("", "", "abc"))
assert.fails(lambda: "abc".partition(""), "empty separator")
assert.fails(lambda: "abc".rpartition(""), "empty separator")

# split and rsplit differ only when maxsplit limits the splits.
assert.eq("a.b.c.d".split(".", 2), ["a", "b", "c.d"])
assert.eq("a.b.c.d".rsplit(".", 2), ["a.b", "c", "d"])
assert.eq("a.b.c.d".split(".", 0), ["a.b.c.d"])
assert.eq("a.b.c.d".rsplit(".", -1), ["a", "b", "c", "d"])

# A delimiter produces empty fields; whitespace splitting never does.
assert.eq("--aa--bb--".split("-"), ["", "", "aa", "", "bb", "", ""])
assert.eq("--aa--bb--".split("-", 1), ["", "-aa--bb--"])
assert.eq("--aa--bb--".rsplit("-", 1), ["--aa--bb-", ""])
assert.eq("  aa  bb  ".split(), ["aa", "bb"])
assert.eq("  aa  bb  ".split(None, 1), ["aa", "bb  "])
assert.eq("  aa  bb  ".rsplit(None, 1), ["  aa", "bb"])
assert.eq("  ".split(), [])
assert.eq("  ".split("."), ["  "])

# splitlines drops or keeps terminators, and a trailing terminator
# adds no empty line.
assert.eq("".splitlines(), [])
assert.eq("\n".splitlines(), [""])
assert.eq("a\n".splitlines(), ["a"])
assert.eq("\nabc\ndef\n".splitlines(), ["", "abc", "def"])
assert.eq("\nabc\ndef\n".splitlines(True), ["\n", "abc\n", "def\n"])
assert.eq("a\n\nb".splitlines(True), ["a\n", "\n", "b"])

# strip family: no argument, None, and the empty string all mean
# whitespace; otherwise the argument is a cutset of characters.
assert.eq(" \tfoo\n ".strip(""), "foo")
assert.eq("blah.h".strip("b.h"), "la")
assert.eq("blah.h".lstrip("b.h"), "lah.h")
assert.eq("blah.h".rstrip("b.h"), "bla")

# count, find, rfind, startswith, and endswith accept start/end
# indices with slice semantics.
assert.eq("banana".count("a", 2), 2)
assert.eq("banana".count("a", -4, -2), 1)
assert.eq("banana".count("a", 0, -100), 0)
assert.eq("foofoo".find("oo", 2), 4)
assert.eq("foofoo".rfind("oo", 1, 4), 1)
assert.eq("foofoo".find(""), 0)
assert.eq("foofoo".rfind(""), 6)
assert.true("abc".startswith("bc", 1))
assert.true(not "abc".startswith("b", 999))
assert.true("abc".endswith("ab", None, -1))

# startswith and endswith also accept a tuple of candidates (only).
assert.true("abc".startswith(("a", "A")))
assert.true(not "ABC".startswith(("b", "B")))
assert.true("ABC".endswith(("c", "C")))
assert.fails(lambda: "123".startswith((1, 2)), "got int, for element 0")
assert.fails(lambda: "123".startswith(["3"]), "got list")

# join requires every element to be a string.
assert.eq(",".join([]), "")
assert.eq(",".join(("a", "b")), "a,b")
assert.fails(lambda: "".join(None), "got NoneType, want iterable")
assert.fails(lambda: "".join(["one", 2]), "want string, got int")
