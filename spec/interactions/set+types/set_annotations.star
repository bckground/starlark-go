# spec: spec.md#type-values

# set[T] is a type value like list[T], with deep matching.
IntSet = set[int]
assert.eq(type(IntSet), "type")
assert.eq(str(IntSet), "set[int]")
assert.true(isinstance(set([1, 2]), IntSet))
assert.true(isinstance(set(), IntSet))
assert.true(not isinstance(set(["a"]), IntSet))
assert.true(not isinstance([1, 2], IntSet))

# Usable as parameter and return annotations.
def evens(xs: set[int]) -> set[int]:
    return set([x for x in xs if x % 2 == 0])

assert.eq(evens(set([1, 2, 3, 4])), set([2, 4]))
assert.fails(lambda: evens(set(["a"])),
             "does not match the type annotation `set\\[int\\]` for argument `xs`")

# Composes with unions.
def maybe_tags(s: set[str] | None) -> int:
    if s == None:
        return 0
    return len(s)

assert.eq(maybe_tags(None), 0)
assert.eq(maybe_tags(set(["a", "b"])), 2)
assert.fails(lambda: maybe_tags(set([1])), "does not match the type annotation")
