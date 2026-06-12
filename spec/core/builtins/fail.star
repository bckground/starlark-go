# spec: spec.md#built-in-constants-and-functions

# fail aborts execution with an error composed of its arguments.
assert.fails(lambda: fail("boom"), "fail: boom")
assert.fails(lambda: fail("a", 1, [2]), "fail: a 1 \\[2\\]")
assert.fails(lambda: fail(), "fail: ")

# A fail abort unwinds through callers; no expression's value is
# produced.
def inner():
    fail("deep")

def outer():
    inner()
    return "unreached"

assert.fails(outer, "fail: deep")
