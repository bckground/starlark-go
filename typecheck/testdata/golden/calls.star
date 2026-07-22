# Errors in call shape and argument types.

def f(x: int, y: str) -> str:
    return y * x

f(1)
f(1, 2, 3)
f(1, z="a")
f("a", "b")
f(1, y=2)
