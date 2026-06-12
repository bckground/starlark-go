# spec: spec.md#try

# At module level, try turns an unhandled error into a failure that
# terminates module execution.
errs = error_tags("E")

def may_fail()!:
    return errs.E(message="config missing")

config = try may_fail() ### "fail: E"
---
# A module-level try whose call succeeds yields the result.
errs = error_tags("E")

def ok()!:
    return 7

value = try ok()
assert.eq(value, 7)
