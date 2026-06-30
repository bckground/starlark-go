# spec: spec.md#annotation-syntax

# Positional-only parameters compose with annotations: the / marker
# does not affect checking, and annotations do not affect call-site
# binding rules.
def scale(s: str, n: int, /, sep: str = "") -> str:
    return sep.join([s] * n)

assert.eq(scale("ab", 2), "abab")
assert.eq(scale("ab", 2, sep="-"), "ab-ab")

# Annotations on positional-only parameters are enforced.
assert.fails(lambda: scale("ab", "two"),
             "does not match the type annotation `int` for argument `n`")

# Binding rules are still enforced alongside the annotations.
assert.fails(lambda: scale(s="ab", n=2),
             'got a value for positional-only parameter "s"')

# All parameter kinds may be annotated in one signature.
def kitchen_sink(a: int, /, b: str, *rest: int, key: bool = False, **extra: str):
    return (a, b, rest, key, extra)

assert.eq(kitchen_sink(1, "b", 2, 3, key=True, note="n"),
          (1, "b", (2, 3), True, {"note": "n"}))
assert.fails(lambda: kitchen_sink(1, "b", "not int"),
             "does not match the type annotation `int` for argument `rest`")
assert.fails(lambda: kitchen_sink(1, "b", note=5),
             "does not match the type annotation `str` for argument `extra`")
