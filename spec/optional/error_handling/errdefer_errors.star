# requires: defer
# spec: spec.md#errdefer

# errdefer is a static error outside a function.
def noop():
    pass

errdefer noop() ### "errdefer statement not within a function"
---
# errdefer is a static error in a function without the ! marker.
def noop():
    pass

def plain():
    errdefer noop() ### "errdefer statement only allowed in error-returning functions"
