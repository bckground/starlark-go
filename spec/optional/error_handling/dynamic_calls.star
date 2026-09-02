# spec: spec.md#static-validation

# Calls whose target is not statically resolvable to an
# error-returning function - a variable, a parameter, an element, the
# result of another call - are not rejected by the static check; the
# requirement that the call be guarded is enforced at runtime instead.
# The requirement is a property of the call site, not of the outcome:
# an unguarded call to an error-returning value is a failure raised
# before the callee runs, whether or not it would have returned an
# error.

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

# A dynamic call that would have returned no error is still a failure.
def calls_succeeding():
    g = succeeds
    return g()

assert.fails(calls_succeeding, "must be handled with try or catch")

# The callee does not run: the failure is raised at the call site,
# before control enters the error-returning function.
ran = []

def records()!:
    ran.append("entered")
    return "fine"

def calls_records():
    g = records
    return g()

assert.fails(calls_records, "must be handled with try or catch")
assert.eq(ran, [])

# An error that a guarded call does produce is still never dropped.
def drops_on_return():
    f = may_fail
    x = f()
    return "no failure"

assert.fails(drops_on_return, "must be handled with try or catch")

def drops_on_later_call():
    f = may_fail
    g = succeeds
    x = f()
    y = g()
    return "no failure"

assert.fails(drops_on_later_call, "must be handled with try or catch")

# The target need not be a variable: any dynamically dispatched call
# is checked the same way.
table = {"f": may_fail}

def drops_via_element():
    x = table["f"]()
    return "no failure"

assert.fails(drops_via_element, "must be handled with try or catch")

def drops_via_parameter(f):
    x = f()
    return "no failure"

assert.fails(lambda: drops_via_parameter(may_fail), "must be handled with try or catch")

def returns_may_fail():
    return may_fail

def drops_via_call_result():
    x = returns_may_fail()()
    return "no failure"

assert.fails(drops_via_call_result, "must be handled with try or catch")

# An error handled by an enclosing catch block leaves nothing pending.
def handled_by_block():
    f = may_fail
    x = f() catch e:
        recover e.tag
    return x

assert.eq(handled_by_block(), errs.E)
