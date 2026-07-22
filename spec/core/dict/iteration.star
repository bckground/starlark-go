# spec: spec.md#dictionaries

# A dict may not be structurally mutated while it is being iterated,
# whether by a loop or a comprehension.
def insert_during_loop():
    d = {1: 1, 2: 1}
    for k in d:
        d[2 * k] = d[k]

assert.fails(insert_during_loop, "insert.*during iteration")

def delete_during_loop():
    d = {1: 1, 2: 1}
    for k in d:
        d.pop(k)

assert.fails(delete_during_loop, "delete.*during iteration")

def insert_during_comprehension():
    def put(d):
        d[3] = 3

    d = {1: 1, 2: 1}
    _ = [put(d) for x in d]

assert.fails(insert_during_comprehension, "insert.*during iteration")

# Destructuring assignment iterates its operand completely before
# assigning, so this is not a mutation during iteration.
def destructure_then_mutate():
    x = {1: 2, 2: 4}
    a, x[0] = x
    return (a, x)

assert.eq(destructure_then_mutate(), (1, {1: 2, 2: 4, 0: 2}))

# Iterating a dict yields its keys, in insertion order, in every
# iterating construct.
def collect(d):
    keys = []
    for k in d:
        keys.append(k)
    return keys

d = {1: 2, 3: 4}
assert.eq(collect(d), [1, 3])
assert.eq([k for k in d], [1, 3])

def varargs(*args):
    return args

assert.eq(varargs(*{"one": 1}), ("one",))

# A **kwargs parameter receives a fresh dict, not an alias of the
# argument.
def kwargs(**kwargs):
    return kwargs

src = {"one": 1}
got = kwargs(**src)
assert.eq(got, src)
src["two"] = 2
assert.ne(got, src)
