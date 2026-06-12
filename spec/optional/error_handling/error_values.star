# spec: spec.md#error-values

errs = error_tags("NotFound", "IOError")

# Calling a tag produces an error value carrying message, cause, and
# extra.
e = errs.NotFound(message="user 42 not found", extra={"id": 42})
assert.eq(type(e), "error")
assert.eq(e.tag, errs.NotFound)
assert.eq(e.message, "user 42 not found")
assert.eq(e.extra, {"id": 42})
assert.eq(e.cause, None)

# cause chains errors.
inner = errs.IOError(message="disk full")
outer = errs.NotFound(message="lookup failed", cause=inner)
assert.eq(outer.cause, inner)
assert.eq(outer.cause.message, "disk full")

# A bare tag returned from an error-returning function is observed as
# an error value with that tag and the tag's name as message.
def fail_bare()!:
    return errs.NotFound

bare = fail_bare() catch err:
    recover err

assert.eq(type(bare), "error")
assert.eq(bare.tag, errs.NotFound)
assert.eq(bare.message, "NotFound")

# A constructed error value returned from an error-returning function
# is observed as-is.
def fail_rich()!:
    return errs.IOError(message="boom", extra=7)

rich = fail_rich() catch err2:
    recover err2

assert.eq(rich.tag, errs.IOError)
assert.eq(rich.message, "boom")
assert.eq(rich.extra, 7)
