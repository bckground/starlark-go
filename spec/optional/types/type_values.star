# spec: spec.md#type-values

# Parameterized types and unions are first-class values.
IntList = list[int]
MaybeInt = int | None
assert.eq(type(IntList), "type")
assert.eq(type(MaybeInt), "type")
assert.eq(str(IntList), "list[int]")
assert.eq(str(MaybeInt), "int | None")

# They may be used as annotations.
def total(xs: IntList) -> int:
    n = 0
    for x in xs:
        n += x
    return n

assert.eq(total([1, 2, 3]), 6)
assert.fails(lambda: total([1, "two"]), "does not match the type annotation")

def opt(x: MaybeInt) -> MaybeInt:
    return x

assert.eq(opt(1), 1)
assert.eq(opt(None), None)
assert.fails(lambda: opt("s"), "does not match the type annotation")

# Container matching is deep.
def keyed(d: dict[str, int]) -> int:
    return len(d)

assert.eq(keyed({"a": 1}), 1)
assert.fails(lambda: keyed({"a": "b"}), "does not match the type annotation")
assert.fails(lambda: keyed({1: 1}), "does not match the type annotation")

def pair(p: tuple[int, str]) -> str:
    return p[1]

assert.eq(pair((1, "x")), "x")
assert.fails(lambda: pair((1, 2)), "does not match the type annotation")
assert.fails(lambda: pair((1, "x", 3)), "does not match the type annotation")

# Nested containers check recursively.
def nested(m: dict[str, list[int]]) -> int:
    return len(m)

assert.eq(nested({"a": [1]}), 1)
assert.fails(lambda: nested({"a": ["s"]}), "does not match the type annotation")

# tuple[T, ...] matches any length of T.
def sum_all(xs: tuple[int, ...]) -> int:
    n = 0
    for x in xs:
        n += x
    return n

assert.eq(sum_all((1, 2, 3)), 6)
assert.eq(sum_all(()), 0)
assert.fails(lambda: sum_all((1, "a")), "does not match the type annotation `tuple\\[int, ...\\]`")

# The legacy list syntax [T1, T2] denotes a union.
def either(x: [int, str]) -> int:
    return 1

assert.eq(either(1), 1)
assert.eq(either("a"), 1)
assert.fails(lambda: either(None), "does not match the type annotation `int | str`")

# Type values compare by the type they denote, and display
# canonically.
assert.eq(eval_type(int), eval_type(int))
assert.ne(eval_type(int), eval_type(str))
assert.eq(str(eval_type(int)), "int")
assert.eq(str(dict[str, int]), "dict[str, int]")
assert.eq(str(tuple[int, ...]), "tuple[int, ...]")
