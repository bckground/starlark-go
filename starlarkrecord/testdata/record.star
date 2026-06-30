# Tests of the record() and field() builtins.

load("assert.star", "assert")

Rec = record(host=str, port=field(int, 80))

# the type repr is structural; the exported name is used for instances
assert.eq(str(Rec), "record(host=str, port=field(int, 80))")
assert.eq(type(Rec), "record_type")

r = Rec(host="localhost")
assert.eq(type(r), "record")
assert.eq(r.host, "localhost")
assert.eq(r.port, 80)
assert.eq(str(r), 'record[Rec](host="localhost", port=80)')
assert.eq(dir(r), ["host", "port"])

# equality: same type and same field values
assert.eq(r, Rec(host="localhost", port=80))
assert.ne(r, Rec(host="localhost", port=81))

# records of distinct types are never equal, even if structurally identical
Other = record(host=str, port=field(int, 80))
assert.ne(r, Other(host="localhost"))

# construction errors
assert.fails(lambda: Rec(), 'missing field "host"')
assert.fails(lambda: Rec(host=1), 'for field "host"')
assert.fails(lambda: Rec("localhost"), "named arguments only")
assert.fails(lambda: Rec(host="x", extra=1), 'unexpected field "extra"')
assert.fails(lambda: record(x=5), 'invalid type for field "x"')
assert.fails(lambda: field(int, "nope"), "default value")
assert.fails(lambda: r.nope, "no .nope field or method")

# export is set-once: aliasing does not rename the type
Alias = Rec
assert.eq(str(Alias(host="h")), 'record[Rec](host="h", port=80)')

# assigning to a global names the type (export is by first assignment)
def make_rec():
    return record(x=int)

Named = make_rec()
assert.eq(str(Named(x=1)), "record[Named](x=1)")

# a type that is never assigned to a global stays anonymous
def check_anon():
    anon = record(x=int)  # local: no export
    assert.eq(str(anon(x=1)), "record[anon](x=1)")

check_anon()

# record types are usable as type annotations
def serve(r: Rec) -> int:
    return r.port

assert.eq(serve(Rec(host="h", port=8080)), 8080)
assert.fails(lambda: serve({"port": 8080}), "does not match the type annotation `Rec`")
assert.fails(lambda: serve(Other(host="h")), "does not match the type annotation `Rec`")

# field types compose with the full annotation grammar
Flexible = record(id=int | None, tags=list[str])
f1 = Flexible(id=None, tags=["a"])
assert.eq(f1.id, None)
assert.fails(lambda: Flexible(id="x", tags=[]), 'for field "id"')
assert.fails(lambda: Flexible(id=1, tags=[1]), 'for field "tags"')

# nested records as field types
Inner = record(x=int)
Outer = record(inner=Inner)
o = Outer(inner=Inner(x=1))
assert.eq(o.inner.x, 1)
assert.fails(lambda: Outer(inner=1), 'for field "inner"')

# isinstance integration
assert.true(isinstance(r, Rec))
assert.true(not isinstance(r, Other))
assert.true(isinstance(Rec, type))

# records can be dict keys (hash over field values)
d = {Rec(host="a"): 1}
assert.eq(d[Rec(host="a")], 1)

# freezing is recursive and harmless
frozen = Rec(host="f")
assert.eq(frozen.port, 80)
