# Tests of type annotations in parse-only mode: annotations are parsed
# and validated syntactically, but ignored at runtime — no checks occur
# and names within annotations are not even resolved.
# option:typesparseonly

load("assert.star", "assert")

def f(x: some_undefined_name) -> another_undefined_name:
    return x

# no runtime checking: any value is accepted
assert.eq(f(1), 1)
assert.eq(f("a"), "a")
assert.eq(f(None), None)

def g(x: int, *args: str, **kwargs: int) -> str:
    return x

# annotations are ignored even when the names are defined
assert.eq(g(None), None)

# annotated assignments are not checked either
def h():
    y: third_undefined_name = 42
    return y

assert.eq(h(), 42)

x: int = "not checked"
assert.eq(x, "not checked")
