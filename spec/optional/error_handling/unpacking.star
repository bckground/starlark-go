# spec: spec.md#catch

errs = error_tags("Err")

# try and catch are expressions, so their results destructure like
# any other.
def returns_tuple()!:
    return (1, 2, 3)

def relay()!:
    a, b, c = try returns_tuple()
    return (a, b, c)

a, b, c = relay() catch (0, 0, 0)
assert.eq((a, b, c), (1, 2, 3))

# On error, the fallback value is what destructures.
def fails()!:
    return errs.Err

x, y = fails() catch (10, 20)
assert.eq((x, y), (10, 20))

q, r, s = fails() catch e:
    recover (100, 200, 300)

assert.eq((q, r, s), (100, 200, 300))

# A fallback that does not match the target arity is the usual
# sequence-assignment error.
def bad_arity():
    m, n = fails() catch (1, 2, 3)

assert.fails(bad_arity, "too many values to unpack")
