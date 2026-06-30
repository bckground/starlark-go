# spec: spec.md#static-validation

# A direct call to an error-returning function must be handled.
errs = error_tags("E")

def may_fail()!:
    return errs.E

may_fail() ### "call to error-returning function .* must be handled with try or catch"
---
errs = error_tags("E")

def may_fail()!:
    return errs.E

def caller():
    may_fail() ### "call to error-returning function .* must be handled with try or catch"
---
# An error-returning call as an argument is evaluated eagerly and
# must be handled, even inside a defer statement.
errs = error_tags("E")

def may_fail()!:
    return errs.E

def sink(x):
    pass

def caller()!:
    defer sink(may_fail()) ### "call to error-returning function .* must be handled with try or catch"
---
# try and catch require a call that can return an error.
def normal():
    return 1

def caller()!:
    x = try normal() ### "try requires call to error-returning function"
---
def normal():
    return 1

x = normal() catch "default" ### "catch requires call to error-returning function"
---
# try is permitted only inside error-returning functions (except at
# module level).
errs = error_tags("E")

def may_fail()!:
    return errs.E

def plain():
    x = try may_fail() ### "try requires enclosing error-returning function"
---
# recover is permitted only inside a catch block.
def f():
    recover 1 ### "recover statement not within a catch block"
