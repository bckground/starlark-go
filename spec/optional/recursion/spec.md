# Unit: recursion

This unit lifts the core specification's ban on recursion
(spec.md#function-definitions): with it, a function may call itself,
directly or indirectly.

## Recursive calls

In core Starlark, a dynamic call to a function already on the call
stack is an error, guaranteeing termination. With this unit, such
calls are permitted; programs may fail to terminate. (A stack
overflow remains an abort.)
