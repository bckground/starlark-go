# spec: spec.md#annotation-semantics

# A mismatched argument is a failure naming the parameter.
def takes_int(x: int) -> int:
    return x

assert.fails(lambda: takes_int("a"),
             "Value `\"a\"` of type `string` does not match the type annotation `int` for argument `x`")

# A mismatched return value is a failure.
def bad_return(x) -> str:
    return x

assert.fails(lambda: bad_return(5),
             "Value `5` of type `int` does not match the type annotation `str` for return type")

# A mismatched annotated assignment is a failure.
def bad_assign():
    y: int = "oops"

assert.fails(bad_assign,
             "Value `\"oops\"` of type `string` does not match the type annotation `int` for assignment `y`")

# An annotation that does not denote a type is a failure when the def
# executes.
def define_bad():
    not_a_type = 5

    def f(x: not_a_type):
        return x

assert.fails(define_bad, "value of type int is not a type")
