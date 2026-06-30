# Errors in attribute access, indexing, and iteration.

def f(xs: list[int], d: dict[str, int], n: int):
    xs.frobnicate()
    d["a"]
    d[0]
    n[0]
    for x in n:
        pass
