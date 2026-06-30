# Errors in returns, annotated assignments, and parameter defaults.

def f() -> str:
    return 1

def g() -> int:
    return  # implicit None

def h(x: int = "a") -> int:
    y: str = x
    return x
