# spec: spec.md#name-binding-and-variables

# Using a global before its binding statement executes is a dynamic
# error.
def f():
    return g ### "global variable g referenced before assignment"

f()

g = 1
---
# Augmented assignment of an unbound global is such a use.
z += 3 ### "global variable z referenced before assignment"
---
(a) += 5 ### "global variable a referenced before assignment"
---
# Augmented assignment makes the name local to a function, so the
# read fails even though a global of the same name is bound.
x = [1]

def f():
    x += [4] ### "local variable x referenced before assignment"

f()
---
# A module-level binding that shadows a built-in takes effect for the
# whole module: uses before the binding statement refer to the (not
# yet assigned) global, not the built-in.
list = []
assert.eq(type(list), "list")

def use():
    return tuple ### "global variable tuple referenced before assignment"

use()

tuple = ()
