# spec: spec.md#annotation-semantics

# Annotated parameters, returns, and assignments accept matching
# values silently.
def add(x: int, y: int) -> int:
    return x + y

assert.eq(add(1, 2), 3)

def greet(name: str, times: int = 2) -> str:
    return name * times

assert.eq(greet("hi"), "hihi")

def variadic(*args: int, **kwargs: str) -> int:
    return len(args) + len(kwargs)

assert.eq(variadic(1, 2, a="x"), 3)

count: int = 41
count2: int = count + 1
assert.eq(count2, 42)

# None in an annotation accepts exactly None.
def nothing(x: None) -> None:
    return x

assert.eq(nothing(None), None)

# A function with no (or a bare) return returns None, which must
# satisfy the return annotation.
def implicit_none() -> int:
    pass

assert.fails(implicit_none,
             "Value `None` of type `NoneType` does not match the type annotation `int` for return type")

# A float annotation also accepts ints; the reverse does not hold.
def halve(x: float) -> float:
    return x / 2

assert.eq(halve(3.0), 1.5)
assert.eq(halve(3), 1.5)
assert.fails(lambda: halve("3"), "does not match the type annotation `float` for argument `x`")
assert.true(isinstance(1, float))
assert.true(not isinstance(1.0, int))
assert.true(isinstance([1, 2.5], list[float]))

# Keyword-only parameters may be annotated.
def kwonly(*, x: int) -> int:
    return x

assert.eq(kwonly(x=1), 1)
assert.fails(lambda: kwonly(x="a"), "does not match the type annotation `int` for argument `x`")

# A default value is checked when it is bound, not at def time; an
# explicit argument leaves a bad default unobserved.
def bad_default(x: int = "oops") -> int:
    return 0

assert.eq(bad_default(1), 0)
assert.fails(bad_default, "does not match the type annotation `int` for argument `x`")

# Annotations are expressions evaluated when the def executes, in the
# enclosing scope.
MyNumber = int

def uses_alias(x: MyNumber) -> MyNumber:
    return x

assert.eq(uses_alias(7), 7)

def make_checked(t):
    def f(x: t):
        return x

    return f

int_checked = make_checked(int)
assert.eq(int_checked(1), 1)
assert.fails(lambda: int_checked("s"), "does not match the type annotation")
