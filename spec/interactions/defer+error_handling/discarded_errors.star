# spec: spec.md#failures

# An error returned by the function called by a defer statement is
# discarded, like the call's return value: the deferred call still
# runs, and the caller observes neither an error nor a failure.
errs = error_tags("CleanupFailed")

ran = []

def cleanup()!:
    ran.append("cleanup")
    return errs.CleanupFailed

def work():
    defer cleanup()
    return "result"

assert.eq(work(), "result")
assert.eq(ran, ["cleanup"])

# The same holds inside an error-returning caller: the discarded
# error does not become the caller's error.
def outer()!:
    defer cleanup()
    return "ok"

assert.eq(outer() catch "handled", "ok")
assert.eq(ran, ["cleanup", "cleanup"])

# A deferred call may itself handle errors with catch.
log = []

def fragile()!:
    return errs.CleanupFailed

def careful_cleanup():
    log.append(fragile() catch "fallback")

def guarded():
    defer careful_cleanup()
    return "done"

assert.eq(guarded(), "done")
assert.eq(log, ["fallback"])

# Deferred calls run after the return value is computed but before
# the function returns: a deferred mutation of the returned object is
# visible to the caller.
def returns_mutated():
    out = ["body"]
    defer out.append("deferred")
    return out

assert.eq(returns_mutated(), ["body", "deferred"])
