# spec: spec.md#static-validation

# Calls whose target is not statically resolvable to an
# error-returning function - a variable, a parameter, an element - are
# not rejected by the static check; the requirement that an error be
# handled is enforced at runtime instead. An unhandled error from such
# a call is a failure: it is never dropped, leaving the call to
# evaluate to None.

errs = error_tags("E")

def may_fail()!:
    return errs.E(message="dropped")

def succeeds()!:
    return "fine"

# Handling a dynamic call with catch or try works as for a direct one.
def catches():
    f = may_fail
    return f() catch "fallback"

assert.eq(catches(), "fallback")

def tries()!:
    f = may_fail
    return try f()

assert.eq(tries() catch "propagated", "propagated")

# A dynamic call that returns no error is unaffected by the check.
def calls_succeeding():
    g = succeeds
    return g()

assert.eq(calls_succeeding(), "fine")

# An unhandled error from a dynamic call is a failure, both when the
# enclosing function returns...
def drops_on_return():
    f = may_fail
    x = f()
    return "no failure"

assert.fails(drops_on_return, "E: dropped")

# ...and when a later call in the same function would overwrite it.
def drops_on_later_call():
    f = may_fail
    g = succeeds
    x = f()
    y = g()
    return "no failure"

assert.fails(drops_on_later_call, "E: dropped")

# The target need not be a variable: any dynamically dispatched call
# is checked the same way.
table = {"f": may_fail}

def drops_via_element():
    x = table["f"]()
    return "no failure"

assert.fails(drops_via_element, "E: dropped")

# The check is about the error channel, not the call: an error handled
# by an enclosing catch block leaves nothing pending.
def handled_by_block():
    f = may_fail
    x = f() catch e:
        recover e.tag
    return x

assert.eq(handled_by_block(), errs.E)
