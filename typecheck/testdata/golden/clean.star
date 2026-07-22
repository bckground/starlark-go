# A well-typed module: the checker must report nothing.

def fib(i: int) -> int:
    if i < 2:
        return i
    return fib(i - 1) + fib(i - 2)

def greeting(name: str, excited: bool = False) -> str:
    suffix = "!" if excited else "."
    return "Hello, " + name + suffix

def total(xs: list[int]) -> int:
    n = 0
    for x in xs:
        n += x
    return n

pairs = list(enumerate(sorted(["b", "a"])))
table: dict[str, int] = {"one": 1}
counts = [total([1, 2, 3]), fib(10)]
maybe: int | None = None
