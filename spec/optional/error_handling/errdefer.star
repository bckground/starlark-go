# requires: defer
# spec: spec.md#errdefer

errs = error_tags("E")

def may_fail(flag)!:
    if flag:
        return errs.E
    return "ok"

# errdefer runs only when the function exits with an error.
def transact(flag, log)!:
    errdefer log.append("rollback")
    defer log.append("close")
    x = try may_fail(flag)
    return x

err_log = []
assert.eq(transact(True, err_log) catch "handled", "handled")
assert.eq(err_log, ["rollback", "close"])

ok_log = []
assert.eq(transact(False, ok_log) catch "handled", "ok")
assert.eq(ok_log, ["close"])

# On an error exit, errdefers run before defers, each in LIFO order.
def ordered(log)!:
    defer log.append("d1")
    errdefer log.append("e1")
    defer log.append("d2")
    errdefer log.append("e2")
    return errs.E

order = []
ordered(order) catch "handled"
assert.eq(order, ["e2", "e1", "d2", "d1"])

# errdefer also runs when an error leaves via return of an error
# value (not only via try).
def direct(log)!:
    errdefer log.append("ran")
    return errs.E(message="m")

direct_log = []
direct(direct_log) catch "handled"
assert.eq(direct_log, ["ran"])

# An error handled inside the function does not trigger errdefer.
def handled_inside(log)!:
    errdefer log.append("rollback")
    x = may_fail(True) catch "fallback"
    return x

inside_log = []
assert.eq(handled_inside(inside_log) catch "outer", "fallback")
assert.eq(inside_log, [])
