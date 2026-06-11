# Errors in unary, binary, and comparison operators.

def f(x: int, s: str):
    a = x + s
    b = -s
    c = x == s          # pointless comparison
    d = x < s
    e = x in s
    ok1 = x == x
    ok2 = s == None or s == "a"

def narrowing_ok(v: int | None) -> bool:
    return v == None
