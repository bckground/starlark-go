# spec: spec.md#error-tags

# error_tags creates a tag set with one attribute per name.
errs = error_tags("NotFound", "Timeout")
assert.eq(type(errs), "error_tags")
assert.eq(type(errs.NotFound), "error_tag")
assert.eq(str(errs.NotFound), "NotFound")
assert.eq(str(errs.Timeout), "Timeout")
assert.eq(type(error_tags()), "error_tags")

# Tags are equal only to themselves.
assert.eq(errs.NotFound, errs.NotFound)
assert.ne(errs.NotFound, errs.Timeout)

# Each error_tags call mints fresh tags, even under the same name.
other = error_tags("NotFound")
assert.ne(errs.NotFound, other.NotFound)

# Accessing an undeclared tag is an error.
assert.fails(lambda: errs.Missing, "has no attribute")

# Tag sets merge with | and +; the merged set contains the original
# tags themselves.
more = error_tags("Conflict")
merged = errs | more
assert.eq(type(merged), "error_tags")
assert.eq(merged.NotFound, errs.NotFound)
assert.eq(merged.Conflict, more.Conflict)
plus = errs + more
assert.eq(plus.Timeout, errs.Timeout)
