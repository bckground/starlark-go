# spec: spec.md#catch

errs = error_tags("E")

def may_fail()!:
    return errs.E

# The catch block introduces a new scope: the error variable shadows
# an enclosing binding and leaves it unchanged.
def shadows():
    e = "outer"
    result = may_fail() catch e:
        recover "caught " + str(e)
    return (result, e)

assert.eq(shadows(), ("caught E", "outer"))

# Names bound inside the block are local to it; enclosing bindings of
# the same name are unchanged.
def block_local():
    seen = "outer"
    result = may_fail() catch e:
        seen = "inner"
        recover seen
    return (result, seen)

assert.eq(block_local(), ("inner", "outer"))

# The block may mutate enclosing values.
def mutates():
    log = []
    result = may_fail() catch e:
        log.append(str(e))
        recover "handled"
    return (result, log)

assert.eq(mutates(), ("handled", ["E"]))

# Nested catch blocks shadow independently.
def nested():
    result = may_fail() catch e:
        inner = may_fail() catch e:
            recover "inner " + str(e)
        recover inner + ", outer " + str(e)
    return result

assert.eq(nested(), "inner E, outer E")

# The error variable may shadow a parameter without affecting it.
def with_param(e)!:
    result = may_fail() catch e:
        recover "caught"
    return result + " param " + e

assert.eq(with_param("p") catch "outer", "caught param p")
