# spec: spec.md#execution-guarantees

# Deferred calls run even when the function exits via an abort, in
# every frame being unwound; the abort then continues.
log = []

def inner():
    defer log.append("inner-cleanup")
    fail("boom")

def outer():
    defer log.append("outer-cleanup")
    inner()

assert.fails(outer, "fail: boom")
assert.eq(log, ["inner-cleanup", "outer-cleanup"])

# The same holds for implicit aborts.
log2 = []

def divides():
    defer log2.append("cleanup")
    return 1 // 0

assert.fails(divides, "floored division by zero")
assert.eq(log2, ["cleanup"])
