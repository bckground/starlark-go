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

# Calling a tag with no arguments yields an error whose message
# defaults to the tag's name.
bare_err = errs.NotFound()
assert.eq(type(bare_err), "error")
assert.eq(bare_err.tag, errs.NotFound)
assert.eq(bare_err.message, "NotFound")
assert.eq(bare_err.cause, None)
assert.eq(bare_err.extra, None)

# extra carries any value.
assert.eq(errs.NotFound(extra=[1, "two"]).extra, [1, "two"])
assert.eq(errs.NotFound(extra={"id": 42}).extra, {"id": 42})

# str of an error is its tag's name, not its message.
assert.eq(str(errs.NotFound(message="custom message")), "NotFound")

# Error tags and error values are false in a Boolean context.
assert.true(not errs.NotFound)
assert.true(not errs.NotFound(message="still false"))

# cause chains errors.
inner = errs.IOError(message="disk full")
outer = errs.NotFound(message="lookup failed", cause=inner)
assert.eq(outer.cause, inner)
assert.eq(outer.cause.message, "disk full")
assert.eq(outer.cause.cause, None)

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
