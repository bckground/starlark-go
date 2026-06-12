# spec: spec.md#defer-statements

# A deferred call runs when the enclosing function exits.
def basic():
    out = []
    defer out.append("deferred")
    out.append("body")
    return out

assert.eq(basic(), ["body", "deferred"])

# Deferred calls run in LIFO order.
def lifo():
    out = []
    defer out.append(1)
    defer out.append(2)
    defer out.append(3)
    return out

assert.eq(lifo(), [3, 2, 1])

# Arguments are evaluated when the defer statement executes.
def eager_args():
    out = []
    i = 0
    defer out.append(i)  # captures 0
    i = 1
    defer out.append(i)  # captures 1
    i = 2
    return out

assert.eq(eager_args(), [1, 0])

# Deferred calls run on early returns too.
def early(flag):
    out = []
    defer out.append("cleanup")
    if flag:
        return out
    out.append("late")
    return out

assert.eq(early(True), ["cleanup"])
assert.eq(early(False), ["late", "cleanup"])

# A defer in a loop schedules one call per iteration.
def loop():
    out = []
    for i in range(3):
        defer out.append(i)
    return out

assert.eq(loop(), [2, 1, 0])

# Nested functions have independent defer stacks.
def nested():
    out = []
    defer out.append("outer-defer")

    def inner():
        defer out.append("inner-defer")
        out.append("inner-body")

    inner()
    out.append("outer-body")
    return out

assert.eq(nested(), ["inner-body", "inner-defer", "outer-body", "outer-defer"])

# The deferred call's return value is discarded.
def discard():
    out = []

    def value():
        out.append("ran")
        return "ignored"

    defer value()
    return out

assert.eq(discard(), ["ran"])
