# spec: spec.md#try

errs = error_tags("E")

def may_fail(flag)!:
    if flag:
        return errs.E
    return "value"

# try yields the call's result on success.
def pass_through(flag)!:
    got = try may_fail(flag)
    return "saw " + got

assert.eq(pass_through(False) catch "handled", "saw value")

# On error, try makes the enclosing function return the error to its
# caller; the rest of the function does not execute.
trace = []

def propagate(flag)!:
    got = try may_fail(flag)
    trace.append("after try")
    return got

assert.eq(propagate(True) catch "handled", "handled")
assert.eq(trace, [])
assert.eq(propagate(False) catch "handled", "value")
assert.eq(trace, ["after try"])

# Errors propagate through a chain of trys, preserving the error
# value.
def level1()!:
    return errs.E(message="deep")

def level2()!:
    return try level1()

def level3()!:
    return try level2()

caught = level3() catch e:
    recover e

assert.eq(caught.tag, errs.E)
assert.eq(caught.message, "deep")

# try may be used in any expression position.
def double(flag)!:
    return (try may_fail(flag)) + "!"

assert.eq(double(False) catch "handled", "value!")
