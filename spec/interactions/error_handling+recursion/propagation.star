# requires: defer
# spec: spec.md#try

errs = error_tags("Negative", "Bottom")

# try propagates an error up an arbitrarily deep recursive stack.
def fact(n)!:
    if n < 0:
        return errs.Negative
    if n == 0:
        return 1
    return n * (try fact(n - 1))

assert.eq(fact(5) catch "error", 120)
assert.eq(fact(-1) catch "error", "error")

# The error surfaces from the depth at which it arose, unwinding
# every intermediate frame.
def descend(n)!:
    if n == 0:
        return errs.Bottom(message="hit bottom", extra=n)
    return try descend(n - 1)

caught = descend(10) catch e:
    recover e

assert.eq(caught.tag, errs.Bottom)
assert.eq(caught.message, "hit bottom")
assert.eq(caught.extra, 0)

# Each recursive frame's errdefer runs as the error unwinds through
# it, innermost first.
unwound = []

def watched(n)!:
    errdefer unwound.append(n)
    if n == 3:
        return errs.Bottom
    return try watched(n + 1)

watched(0) catch "handled"
assert.eq(unwound, [3, 2, 1, 0])

# An intermediate frame may intercept the error and stop the unwind.
def guarded(n)!:
    if n == 0:
        return errs.Bottom
    if n == 2:
        result = guarded(n - 1) catch "stopped at 2"
        return result
    return try guarded(n - 1)

assert.eq(guarded(4) catch "outer", "stopped at 2")
