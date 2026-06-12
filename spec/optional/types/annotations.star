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
