# spec: spec.md#recover

errs = error_tags("E")

def may_fail()!:
    return errs.E

# recover ends the catch block; its operand becomes the value of the
# catch expression, and execution resumes after the block.
def with_value():
    trace = []
    x = may_fail() catch e:
        trace.append("handling")
        recover "recovered"
    trace.append("after")
    return (x, trace)

assert.eq(with_value(), ("recovered", ["handling", "after"]))

# Statements in the block after recover do not execute.
def stops():
    trace = []
    x = may_fail() catch e:
        recover "r"
        trace.append("unreached")
    return trace

assert.eq(stops(), [])

# The recover operand may use the error value.
def uses_error():
    x = may_fail() catch e:
        recover e.message + "!"
    return x

assert.eq(uses_error(), "E!")

# recover clears the error: the enclosing error-returning function
# continues normally.
def clears()!:
    x = may_fail() catch e:
        recover "ok"
    return x + " then some"

assert.eq(clears() catch "outer", "ok then some")
