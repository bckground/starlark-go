# spec: spec.md#function-definitions

# The recursion check detects re-entry of the same function body, not
# merely the same function value: a fresh closure per step does not
# evade it.
def make_fib():
    def fib(go, x):
        return x if x < 2 else go(go, x - 1) + go(go, x - 2)

    return fib

assert.fails(lambda: make_fib()(make_fib(), 10), "called recursively")

# Consequently a helper applied to a function that itself uses the
# helper is rejected, even though the program is bounded.
def map(f, seq):
    return [f(x) for x in seq]

def double(x):
    return x + x

assert.eq(map(double, [1, 2, 3]), [2, 4, 6])

def mapdouble(x):
    return map(double, x)

assert.fails(lambda: map(mapdouble, ([1, 2], [3])),
             "function map called recursively")

# Mutually recursive definitions are fine as long as no call cycle
# occurs dynamically.
calls = []

def yin(x):
    calls.append("yin")
    if x:
        yang(False)

def yang(x):
    calls.append("yang")
    if x:
        yin(False)

yin(True)
assert.eq(calls, ["yin", "yang"])
calls.clear()
yang(True)
assert.eq(calls, ["yang", "yin"])
