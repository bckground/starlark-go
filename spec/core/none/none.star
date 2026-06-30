# spec: spec.md#none

# None is a distinguished value used to indicate absence.
assert.eq(type(None), "NoneType")
assert.eq(None, None)
assert.ne(None, False)
assert.ne(None, 0)
assert.ne(None, "")

# None is falsy.
assert.true(not None)
assert.eq(bool(None), False)

# A function with no return statement, or a bare return, returns None.
def implicit():
    pass

def bare():
    return

assert.eq(implicit(), None)
assert.eq(bare(), None)

# None is not ordered.
assert.fails(lambda: None < None, "NoneType < NoneType not implemented")
