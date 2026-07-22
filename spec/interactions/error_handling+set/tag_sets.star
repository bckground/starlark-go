# spec: spec.md#error-tags

# Error tags are hashable, so they work as set elements: collecting
# them deduplicates by tag identity.
errs = error_tags("NotFound", "Timeout", "Conflict")

seen = set([errs.NotFound, errs.Timeout, errs.NotFound])
assert.eq(len(seen), 2)
assert.true(errs.NotFound in seen)
assert.true(errs.Conflict not in seen)

# Same-named tags from different sets are distinct elements.
other = error_tags("NotFound")
both = set([errs.NotFound, other.NotFound])
assert.eq(len(both), 2)

# A set of tags classifies caught errors.
RETRIABLE = set([errs.Timeout, errs.Conflict])

def may_fail(tag)!:
    return tag

def classify(tag):
    result = may_fail(tag) catch e:
        if e.tag in RETRIABLE:
            recover "retry"
        recover "give up"
    return result

assert.eq(classify(errs.Timeout), "retry")
assert.eq(classify(errs.Conflict), "retry")
assert.eq(classify(errs.NotFound), "give up")

# Collecting the distinct tags observed across many failures.
observed = set()

def record(tag):
    may_fail(tag) catch e:
        observed.add(e.tag)
        recover None

record(errs.Timeout)
record(errs.Timeout)
record(errs.NotFound)
assert.eq(len(observed), 2)
