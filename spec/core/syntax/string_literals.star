# spec: spec.md#string-literals

# Single and double quotes are equivalent.
assert.eq('hello', "hello")

# Escape sequences.
assert.eq("\n", chr(10))
assert.eq("\t", chr(9))
assert.eq("\\", chr(92))
assert.eq("\"", chr(34))
assert.eq('\'', chr(39))
assert.eq("\x41", "A")    # hexadecimal escape
assert.eq("\101", "A")    # octal escape
assert.eq("\u0041", "A")  # 16-bit Unicode escape
assert.eq("\U0001F600", chr(0x1F600))  # 32-bit Unicode escape

# Raw strings do not process escape sequences.
assert.eq(r"a\nb", "a" + chr(92) + "nb")
assert.eq(len(r"\t"), 2)

# Triple-quoted strings may span multiple lines.
s = """line one
line two"""
assert.eq(s, "line one" + chr(10) + "line two")
assert.eq('''abc''', "abc")
