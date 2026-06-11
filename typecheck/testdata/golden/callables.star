# Errors in higher-order callable annotations.

def apply(f: typing.Callable[[int], str], x: int) -> str:
    return f(x)

def good(i: int) -> str:
    return str(i)

def bad_param(s: str) -> str:
    return s

def bad_result(i: int) -> int:
    return i

a = apply(good, 1)
b = apply(bad_param, 1)
c = apply(bad_result, 1)
