# spec: spec.md#defer-statements

# defer may not appear at module level.
defer print(1) ### "defer statement not within a function"
