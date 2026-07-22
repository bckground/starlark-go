# spec: spec.md#recursive-calls

# Direct recursion.
def fact(n):
    return 1 if n <= 1 else n * fact(n - 1)

assert.eq(fact(5), 120)
assert.eq(fact(0), 1)

# Mutual recursion.
def is_even(n):
    return True if n == 0 else is_odd(n - 1)

def is_odd(n):
    return False if n == 0 else is_even(n - 1)

assert.true(is_even(10))
assert.true(is_odd(7))

# Recursion through a value.
def apply_twice(f, x):
    return f(f(x))

def tree_sum(t):
    return t if type(t) == "int" else sum_list([tree_sum(x) for x in t])

def sum_list(l):
    total = 0
    for x in l:
        total += x
    return total

assert.eq(tree_sum([1, [2, [3]], 4]), 10)
assert.eq(apply_twice(fact, 3), 720)
