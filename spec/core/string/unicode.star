# spec: spec.md#strings

# A string is a sequence of bytes, usually UTF-8 text: len counts
# bytes, and indexing can land mid-encoding.
assert.eq(len("Hello, 世界!"), 14)
assert.eq(len("𐐷"), 4)  # U+10437: 4 bytes of UTF-8

# chr returns the UTF-8 encoding of a code point; ord is its inverse.
assert.eq(chr(65), "A")        # 1-byte encoding
assert.eq(chr(1049), "Й")      # 2-byte encoding
assert.eq(chr(0x4E16), "世")   # 3-byte encoding
assert.eq(chr(0x1F63F), "😿")  # 4-byte encoding
assert.eq(ord("Й"), 0x419)
assert.eq(ord("😿"), 0x1F63F)
assert.fails(lambda: chr(-1), "Unicode code point -1 out of range")
assert.fails(lambda: chr(0x110000), "out of range")
assert.fails(lambda: ord(""), "string encodes 0 Unicode code points, want 1")

# Decoding invalid text yields U+FFFD, the replacement character.
assert.eq(ord("Й"[1:]), 0xFFFD)

# elems/elem_ords yield the bytes; codepoints/codepoint_ords decode
# UTF-8.
s = "aЙ😿"
assert.eq(list(s.codepoints()), ["a", "Й", "😿"])
assert.eq(list(s.codepoint_ords()), [97, 1049, 128575])
assert.eq(list(s.elem_ords()), [97, 208, 153, 240, 159, 152, 191])
assert.eq(len(list(s.elems())), 7)
assert.eq(list("".codepoints()), [])
assert.eq(list("".elems()), [])

# A truncated encoding decodes as replacement characters, one per
# undecodable byte.
assert.eq(list(("A" + "😿Z"[1:]).codepoints()), ["A", "�", "�", "�", "Z"])

# Case conversions are Unicode-aware, including multibyte text and
# the titlecase digraphs.
assert.eq("por qué".upper(), "POR QUÉ")
assert.eq("¿Por qué?".lower(), "¿por qué?")
assert.eq("hElLo, WoRlD!".title(), "Hello, World!")
assert.eq("por qué".title(), "Por Qué")
assert.eq("ǉubović".upper(), "ǇUBOVIĆ")
assert.eq("ǇUBOVIĆ".lower(), "ǉubović")
assert.eq("ǉubović".title(), "ǈubović")  # title case, not upper case
assert.true("ǅenan ǈubović".istitle())
assert.true(not "Ǆenan Ǉubović".istitle())
assert.true("ǆenan ǉubović".islower())

# capitalize uppercases the first code point and lowercases the rest.
assert.eq("hElLo, WoRlD!".capitalize(), "Hello, world!")
assert.eq("por qué".capitalize(), "Por qué")
assert.eq("¿Por qué?".capitalize(), "¿por qué?")  # ¿ has no upper case

# The is* predicates classify the whole string.
def check_predicates():
    predicates = ["alnum", "alpha", "digit", "lower", "space", "title", "upper"]
    table = {
        "Hello, World!": "title",
        "hello, world!": "lower",
        "base64": "alnum lower",
        "HAL-9000": "upper",
        "Catch-22": "title",
        "": "",
        "\n\t\r": "space",
        "abc": "alnum alpha lower",
        "ABC": "alnum alpha upper",
        "123": "alnum digit",
        "ǅ ǈ": "title",
    }
    for s, want in table.items():
        got = " ".join([name for name in predicates if getattr(s, "is" + name)()])
        if got != want:
            assert.fail("%r matched [%s], want [%s]" % (s, got, want))

check_predicates()
