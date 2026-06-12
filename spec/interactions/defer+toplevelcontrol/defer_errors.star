# spec: spec.md#defer-statements

# defer requires a function even where toplevelcontrol permits
# control statements at module level: a top-level if or for body is
# still not a function.
if True:
    defer print(1) ### "defer statement not within a function"
---
for x in [1]:
    defer print(x) ### "defer statement not within a function"
