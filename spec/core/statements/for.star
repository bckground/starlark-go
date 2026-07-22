# spec: spec.md#for-statements

# A for statement iterates an iterable sequence: list, tuple, dict
# (its keys), or range.
def collect(iterable):
    out = []
    for x in iterable:
        out.append(x)
    return out

assert.eq(collect([1, 2, 3]), [1, 2, 3])
assert.eq(collect((1, 2)), [1, 2])
assert.eq(collect({"a": 1, "b": 2}), ["a", "b"])
assert.eq(collect(range(3)), [0, 1, 2])

# Non-iterable operands are an error.
assert.fails(lambda: collect(5), "int value is not iterable")
assert.fails(lambda: collect("abc"), "string value is not iterable")
assert.fails(lambda: collect(None), "NoneType value is not iterable")

# The loop variables may destructure each element.
def items(d):
    out = []
    for k, v in d.items():
        out.append(k + str(v))
    return out

assert.eq(items({"a": 1, "b": 2}), ["a1", "b2"])

# Destructuring nests, and any assignable target may appear,
# including an element of a collection.
def nested():
    res = []
    for (x, y), z in [(["a", "b"], 3), (["c", "d"], 4)]:
        res.append((x, y, z))
    return res

assert.eq(nested(), [("a", "b", 3), ("c", "d", 4)])

def element_target():
    a = {}
    for i, a[i] in [("one", 1), ("two", 2)]:
        pass
    return a

assert.eq(element_target(), {"one": 1, "two": 2})

# break exits the nearest enclosing loop; continue skips to the next
# iteration.
def search(haystack, needle):
    found = -1
    for i in range(len(haystack)):
        if haystack[i] == needle:
            found = i
            break
    return found

assert.eq(search([5, 6, 7], 6), 1)
assert.eq(search([5, 6, 7], 9), -1)

def evens(seq):
    out = []
    for x in seq:
        if x % 2 != 0:
            continue
        out.append(x)
    return out

assert.eq(evens([1, 2, 3, 4]), [2, 4])

def nested_break():
    out = []
    for x in range(3):
        for y in range(3):
            if y > x:
                break
            out.append((x, y))
    return out

assert.eq(nested_break(), [(0, 0), (1, 0), (1, 1), (2, 0), (2, 1), (2, 2)])
