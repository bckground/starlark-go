# spec: spec.md#statements

# In core Starlark, control statements may not appear at module
# level; they must be within a function.
if True: ### "if statement not within a function"
    pass
---
for x in [1]: ### "for loop not within a function"
    pass
---
# return is permitted only within a function.
return 1 ### "return statement not within a function"
---
# break and continue are permitted only within a loop.
def f():
    break ### "break not in a loop"
---
def f():
    continue ### "continue not in a loop"
---
def f():
    if True:
        break ### "break not in a loop"
