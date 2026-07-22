# Tests of Starlark type annotations (runtime checking enabled).
# option:types option:set

load("assert.star", "assert")
load("typing.star", "typing")

# basic parameter checks
def double(x: int) -> int:
    return 2 * x

assert.eq(double(21), 42)
assert.fails(lambda: double("a"), "Value `\"a\"` of type `string` does not match the type annotation `int` for argument `x`")

# str display name matches values of type "string"
def greet(name: str) -> str:
    return "hello " + name

assert.eq(greet("world"), "hello world")
assert.fails(lambda: greet(42), "does not match the type annotation `str` for argument `name`")

# a float annotation also accepts ints (numeric coercion, rust parity);
# the reverse does not hold
def halve(x: float) -> float:
    return x / 2

assert.eq(halve(3.0), 1.5)
assert.eq(halve(3), 1.5)
assert.fails(lambda: halve("3"), "does not match the type annotation `float` for argument `x`")
assert.true(isinstance(1, float))
assert.true(not isinstance(1.0, int))
assert.true(isinstance([1, 2.5], list[float]))

# defaults flow through the checks; annotated defaults are checked when bound
def pad(s: str, n: int = 4) -> str:
    return s + " " * n

assert.eq(pad("x", 0), "x")
assert.fails(lambda: pad("x", "y"), "does not match the type annotation `int` for argument `n`")

# a bad default value fails when the default is used...
def bad_default(x: int = "oops") -> int:
    return 0

assert.fails(lambda: bad_default(), "does not match the type annotation `int` for argument `x`")
# ...but not when an explicit argument is given
assert.eq(bad_default(1), 0)

# keyword-only parameters
def kwonly(*, x: int) -> int:
    return x

assert.eq(kwonly(x=1), 1)
assert.fails(lambda: kwonly(x="a"), "does not match the type annotation `int` for argument `x`")

# *args: each element is checked against the annotation
def sum_args(*args: int) -> int:
    total = 0
    for a in args:
        total += a
    return total

assert.eq(sum_args(1, 2, 3), 6)
assert.eq(sum_args(), 0)
assert.fails(lambda: sum_args(1, "two"), "does not match the type annotation `int` for argument `args`")

# **kwargs: each value is checked against the annotation
def collect(**kwargs: str) -> list:
    return sorted(kwargs.values())

assert.eq(collect(a="x", b="y"), ["x", "y"])
assert.fails(lambda: collect(a=1), "does not match the type annotation `str` for argument `kwargs`")

# return types, including implicit return None
def implicit_none() -> int:
    pass

assert.fails(implicit_none, "Value `None` of type `NoneType` does not match the type annotation `int` for return type")

def returns_none() -> None:
    pass

assert.eq(returns_none(), None)

# parameterized and union types
def head(xs: list[int]) -> int | None:
    return xs[0] if xs else None

assert.eq(head([1, 2]), 1)
assert.eq(head([]), None)
assert.fails(lambda: head([1, "a"]), "does not match the type annotation `list\\[int\\]` for argument `xs`")

def lookup(d: dict[str, int], key: str) -> int:
    return d[key]

assert.eq(lookup({"a": 1}, "a"), 1)
assert.fails(lambda: lookup({"a": "b"}, "a"), "does not match the type annotation `dict\\[str, int\\]` for argument `d`")

def swap(pair: tuple[int, str]) -> tuple[str, int]:
    return (pair[1], pair[0])

assert.eq(swap((1, "a")), ("a", 1))
assert.fails(lambda: swap(("a", 1)), "does not match the type annotation `tuple\\[int, str\\]`")

def total(xs: tuple[int, ...]) -> int:
    n = 0
    for x in xs:
        n += x
    return n

assert.eq(total((1, 2, 3)), 6)
assert.eq(total(()), 0)
assert.fails(lambda: total((1, "a")), "does not match the type annotation `tuple\\[int, ...\\]`")

def elems(s: set[int]) -> list:
    return sorted(list(s))

assert.eq(elems(set([2, 1])), [1, 2])
assert.fails(lambda: elems(set(["a"])), "does not match the type annotation `set\\[int\\]`")

# typing module
def apply(f: typing.Callable[[int], int], x: int) -> int:
    return f(x)

assert.eq(apply(double, 3), 6)
assert.fails(lambda: apply(1, 3), "does not match the type annotation `typing.Callable\\[\\[int\\], int\\]`")

def consume(xs: typing.Iterable) -> int:
    n = 0
    for x in xs:
        n += 1
    return n

assert.eq(consume([1, 2]), 2)
assert.eq(consume((1, 2, 3)), 3)
assert.fails(lambda: consume(1), "does not match the type annotation `typing.Iterable`")

def anything(x: typing.Any) -> typing.Any:
    return x

assert.eq(anything("ok"), "ok")
assert.eq(anything(None), None)

def nothing(x: typing.Never) -> int:
    return 0

assert.fails(lambda: nothing(1), "does not match the type annotation `typing.Never`")

# legacy list-union syntax [T1, T2]
def either(x: [int, str]) -> int:
    return 1

assert.eq(either(1), 1)
assert.eq(either("a"), 1)
assert.fails(lambda: either(None), "does not match the type annotation `int | str`")

---
# annotated assignments
# option:types

load("assert.star", "assert")

x: int = 1
assert.eq(x, 1)

def f():
    y: str = "a"
    z: int | None = None
    t: list[int] = [1, 2]
    return (y, z, t)

assert.eq(f(), ("a", None, [1, 2]))

def g():
    y: int = "oops" ### "Value `\"oops\"` of type `string` does not match the type annotation `int` for assignment `y`"
    return y

g()

---
# annotated assignment checks each execution
# option:types

load("assert.star", "assert")

def loop_check():
    for v in [1, 2, "three"]:
        x: int = v ### "does not match the type annotation `int` for assignment `x`"

loop_check()

---
# a runtime annotation that is not a type fails at def time
# option:types

T = 5

def f(x: T): ### "invalid type annotation for parameter x of function f: value of type int is not a type"
    pass

---
# string values are not types (rust parity)
# option:types

T = "int"

def f(x: T): ### "string literal is not allowed in type expression"
    pass

---
# type aliases via ordinary assignment work
# option:types

load("assert.star", "assert")

IntList = list[int]
Maybe = int | None

def f(xs: IntList, m: Maybe) -> int:
    return len(xs)

assert.eq(f([1], None), 1)
assert.eq(f([], 2), 0)
assert.fails(lambda: f(["a"], None), "does not match the type annotation `list\\[int\\]`")
assert.fails(lambda: f([], "x"), "does not match the type annotation `int | None`")

---
# annotations are evaluated in the enclosing scope at def time, like defaults
# option:types

load("assert.star", "assert")

def make(t):
    def inner(x: t) -> t:
        return x
    return inner

int_id = make(int)
str_id = make(str)
assert.eq(int_id(1), 1)
assert.eq(str_id("a"), "a")
assert.fails(lambda: int_id("a"), "does not match the type annotation `int`")
assert.fails(lambda: str_id(1), "does not match the type annotation `str`")

---
# interaction with error-returning (!) functions:
# an error return bypasses the return type check
# option:types

load("assert.star", "assert")

errors = error_tags("NotFound")

def find(id: int)! -> str:
    if id < 0:
        return errors.NotFound
    return "user_" + str(id)

assert.eq(find(42) catch "guest", "user_42")
assert.eq(find(-1) catch "guest", "guest")

# but a non-error return value of a ! function is checked
def bad()! -> int:
    return "nope"

def call_bad()!:
    return try bad()

assert.fails(lambda: call_bad() catch "x", "Value `\"nope\"` of type `string` does not match the type annotation `int` for return type")

---
# types as first-class values
# option:types

load("assert.star", "assert")
load("typing.star", "typing")

assert.eq(eval_type(int), eval_type(int))
assert.ne(eval_type(int), eval_type(str))
assert.eq(str(eval_type(int)), "int")
assert.eq(str(list[int]), "list[int]")
assert.eq(str(dict[str, int]), "dict[str, int]")
assert.eq(str(tuple[int, ...]), "tuple[int, ...]")
assert.eq(str(int | None), "int | None")
assert.eq(str(int | str | None), "int | str | None")
assert.eq(str(typing.Any), "typing.Any")
assert.eq(str(typing.Callable[[int], str]), "typing.Callable[[int], str]")
assert.eq(str(typing.Iterable[int]), "typing.Iterable[int]")
assert.eq(type(list[int]), "type")

# unions dedup and flatten
assert.eq(str(int | None | int), "int | None")

# isinstance
assert.true(isinstance(1, int))
assert.true(not isinstance("a", int))
assert.true(isinstance(None, int | None))
assert.true(isinstance([1, 2], list[int]))
assert.true(not isinstance([1, "a"], list[int]))  # deep matching
assert.true(isinstance({"a": 1}, dict[str, int]))
assert.true(not isinstance({"a": "b"}, dict[str, int]))
assert.true(isinstance((1, 2, 3), tuple[int, ...]))
assert.true(isinstance(len, typing.Callable))
assert.true(isinstance(int, type))
assert.true(isinstance(list[str], type))
assert.true(not isinstance(1, type))
assert.fails(lambda: isinstance(1, 2), "isinstance: value of type int is not a type")

# eval_type errors
assert.fails(lambda: eval_type("int"), "string literal is not allowed in type expression")
assert.fails(lambda: eval_type(5), "value of type int is not a type")

# parameterization errors
assert.fails(lambda: dict[int], "dict\\[...\\] expects exactly 2 type arguments, got 1")
assert.fails(lambda: list[int, str], "list\\[...\\] expects exactly 1 type argument, got 2")

---
# shadowing: a user binding of a constructor name is just a value
# option:types

load("assert.star", "assert")

def f():
    int = "shadowed"
    return int

assert.eq(f(), "shadowed")

---
# annotated assignments in loops: constant type expressions are cached
# per function value (TYPEFETCH/TYPESTORE); the behavior must be
# indistinguishable from re-evaluation
# option:types

load("assert.star", "assert")

def cached_ok():
    r = []
    for i in range(3):
        x: int = i
        xs: list[int] = [i]
        u: int | None = i if i % 2 == 0 else None
        t: tuple[int, str] = (i, "a")
        r.append(x)
    return r

assert.eq(cached_ok(), [0, 1, 2])

# repeated calls reuse the same function value (and its cache)
def cached_call(v) -> int:
    y: str | None = v
    return 0

assert.eq(cached_call("s"), 0)
assert.eq(cached_call(None), 0)
assert.fails(lambda: cached_call(1), "does not match the type annotation `str \\| None` for assignment `y`")

# a mismatch on a later iteration is still detected
def cached_fail():
    for v in [1, 2, "boom"]:
        x: int = v

assert.fails(cached_fail, "Value `\"boom\"` of type `string` does not match the type annotation `int` for assignment `x`")

# non-constant type expressions (here: a local alias) are re-evaluated
# on every execution, so rebinding the alias takes effect
def local_alias():
    T = int
    x: T = 1
    T = str
    x: T = "a"
    return x

assert.eq(local_alias(), "a")

def local_alias_fail():
    T = int
    x: T = 1
    T = str
    x: T = 2  # T is now str

assert.fails(local_alias_fail, "does not match the type annotation `str` for assignment `x`")

# closures created in a loop each get their own cache
def make_checkers():
    fns = []
    for i in range(2):
        def check(v) -> bool:
            y: int = v
            return True
        fns.append(check)
    return fns

def check_closures():
    for f in make_checkers():
        assert.true(f(1))
        assert.fails(lambda: f("a"), "does not match the type annotation `int` for assignment `y`")

check_closures()
