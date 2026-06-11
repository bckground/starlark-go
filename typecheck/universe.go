// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typecheck

// Hand-written type signatures for the universal builtins of
// starlark/library.go and the methods of the built-in types.
// starlark-rust derives this information from its #[starlark_module]
// macros; in Go it must be maintained by hand. Anything not listed
// types as an AnyCallable, so omissions cost precision, never
// correctness.

import "maps"

// Signature construction helpers.

func opt(name string, ty Ty) Param {
	return Param{Name: name, Mode: PosOrName, Required: false, Ty: ty}
}
func posOnly(ty Ty) Param    { return Param{Mode: PosOnly, Required: true, Ty: ty} }
func optPosOnly(ty Ty) Param { return Param{Mode: PosOnly, Required: false, Ty: ty} }
func nameOnly(name string, ty Ty) Param {
	return Param{Name: name, Mode: NameOnly, Required: false, Ty: ty}
}
func varArgs(ty Ty) Param   { return Param{Mode: ArgsMode, Ty: ty} }
func varKwargs(ty Ty) Param { return Param{Mode: KwargsMode, Ty: ty} }

func fn(name string, result Ty, params ...Param) Ty {
	return Function(name, &ParamSpec{Params: params}, result)
}

// universeTypes returns the types of the universal environment.
// It is computed once.
var universeTypes = func() map[string]Ty {
	num := Union(Prim("int"), Prim("float"))
	iterAny := Iter(Any())
	str := Prim("string")
	intT := Prim("int")
	boolT := Prim("bool")

	return map[string]Ty{
		"None":  None(),
		"True":  boolT,
		"False": boolT,

		"abs":       fn("abs", num, posOnly(num)),
		"any":       fn("any", boolT, posOnly(iterAny)),
		"all":       fn("all", boolT, posOnly(iterAny)),
		"bool":      fn("bool", boolT, optPosOnly(Any())),
		"bytes":     fn("bytes", Prim("bytes"), posOnly(Any())),
		"chr":       fn("chr", str, posOnly(intT)),
		"dict":      fn("dict", Dict(Any(), Any()), optPosOnly(Any()), varKwargs(Any())),
		"dir":       fn("dir", List(str), posOnly(Any())),
		"enumerate": fn("enumerate", List(Tuple(intT, Any())), posOnly(iterAny), optPosOnly(intT)),
		"fail":      fn("fail", Never(), varArgs(Any()), nameOnly("sep", str)),
		"float":     fn("float", Prim("float"), optPosOnly(Any())),
		"getattr":   fn("getattr", Any(), posOnly(Any()), posOnly(str), optPosOnly(Any())),
		"hasattr":   fn("hasattr", boolT, posOnly(Any()), posOnly(str)),
		"hash":      fn("hash", intT, posOnly(Union(str, Prim("bytes")))),
		"int":       fn("int", intT, optPosOnly(Any()), opt("base", intT)),
		"len":       fn("len", intT, posOnly(Any())),
		"list":      fn("list", List(Any()), optPosOnly(iterAny)),
		"max":       fn("max", Any(), varArgs(Any()), nameOnly("key", Any())),
		"min":       fn("min", Any(), varArgs(Any()), nameOnly("key", Any())),
		"ord":       fn("ord", intT, posOnly(Union(str, Prim("bytes")))),
		"print":     fn("print", None(), varArgs(Any()), nameOnly("sep", str)),
		"range":     fn("range", Prim("range"), posOnly(intT), optPosOnly(intT), optPosOnly(intT)),
		"repr":      fn("repr", str, posOnly(Any())),
		"reversed":  fn("reversed", List(Any()), posOnly(iterAny)),
		"set":       fn("set", Set(Any()), optPosOnly(iterAny)),
		"sorted":    fn("sorted", List(Any()), posOnly(iterAny), nameOnly("key", Any()), nameOnly("reverse", boolT)),
		"str":       fn("str", str, optPosOnly(Any())),
		"tuple":     fn("tuple", AnyTuple(), optPosOnly(iterAny)),
		"type":      fn("type", str, posOnly(Any())),
		"zip":       fn("zip", List(AnyTuple()), varArgs(iterAny)),

		// typing extension
		"isinstance": fn("isinstance", boolT, posOnly(Any()), posOnly(Any())),
		"eval_type":  fn("eval_type", TypeType(), posOnly(Any())),

		// fork-specific error handling
		"error_tags": fn("error_tags", Any(), varArgs(str)),
	}
}()

// UniverseEnv returns an Env describing the universal environment.
func UniverseEnv() Env {
	env := make(Env, len(universeTypes))
	maps.Copy(env, universeTypes)
	return env
}

// Method tables. The list/dict/set methods depend on the receiver's
// element types, so they are functions, not flat maps (this mirrors
// rust's special-casing in TypingOracleCtx).

func listMethod(elem Ty, name string) (Ty, bool) {
	intT := Prim("int")
	switch name {
	case "append":
		return fn("append", None(), posOnly(elem)), true
	case "clear":
		return fn("clear", None()), true
	case "extend":
		return fn("extend", None(), posOnly(Iter(elem))), true
	case "index":
		return fn("index", intT, posOnly(elem), optPosOnly(intT), optPosOnly(intT)), true
	case "insert":
		return fn("insert", None(), posOnly(intT), posOnly(elem)), true
	case "pop":
		return fn("pop", elem, optPosOnly(intT)), true
	case "remove":
		return fn("remove", None(), posOnly(elem)), true
	}
	return Never(), false
}

func dictMethod(key, val Ty, name string) (Ty, bool) {
	switch name {
	case "clear":
		return fn("clear", None()), true
	case "get":
		// get(k) -> V | None; get(k, default) -> V | typeof(default):
		// approximated as the union of both signatures' results.
		return fn("get", Union(val, None(), Any()), posOnly(key), optPosOnly(Any())), true
	case "items":
		return fn("items", List(Tuple(key, val))), true
	case "keys":
		return fn("keys", List(key)), true
	case "pop":
		return fn("pop", Union(val, Any()), posOnly(key), optPosOnly(Any())), true
	case "popitem":
		return fn("popitem", Tuple(key, val)), true
	case "setdefault":
		return fn("setdefault", Union(val, Any()), posOnly(key), optPosOnly(Any())), true
	case "update":
		return fn("update", None(), optPosOnly(Any()), varKwargs(Any())), true
	case "values":
		return fn("values", List(val)), true
	}
	return Never(), false
}

func setMethod(elem Ty, name string) (Ty, bool) {
	boolT := Prim("bool")
	this := Set(elem)
	switch name {
	case "add":
		return fn("add", None(), posOnly(elem)), true
	case "clear":
		return fn("clear", None()), true
	case "difference":
		return fn("difference", this, posOnly(Iter(Any()))), true
	case "discard":
		return fn("discard", None(), posOnly(elem)), true
	case "intersection":
		return fn("intersection", this, posOnly(Iter(Any()))), true
	case "issubset":
		return fn("issubset", boolT, posOnly(Iter(Any()))), true
	case "issuperset":
		return fn("issuperset", boolT, posOnly(Iter(Any()))), true
	case "pop":
		return fn("pop", elem), true
	case "remove":
		return fn("remove", None(), posOnly(elem)), true
	case "symmetric_difference":
		return fn("symmetric_difference", this, posOnly(Iter(Any()))), true
	case "union":
		return fn("union", this, optPosOnly(Iter(Any()))), true
	}
	return Never(), false
}

func stringMethod(name string) (Ty, bool) {
	str := Prim("string")
	intT := Prim("int")
	boolT := Prim("bool")
	listStr := List(str)
	switch name {
	case "capitalize", "lower", "lstrip", "rstrip", "strip", "title", "upper":
		return fn(name, str, optPosOnly(str)), true
	case "count", "find", "index", "rfind", "rindex":
		return fn(name, intT, posOnly(str), optPosOnly(Union(intT, None())), optPosOnly(Union(intT, None()))), true
	case "endswith", "startswith":
		return fn(name, boolT, posOnly(Union(str, AnyTuple())), optPosOnly(intT), optPosOnly(intT)), true
	case "isalnum", "isalpha", "isdigit", "islower", "isspace", "istitle", "isupper":
		return fn(name, boolT), true
	case "join":
		return fn("join", str, posOnly(Iter(str))), true
	case "removeprefix", "removesuffix":
		return fn(name, str, posOnly(str)), true
	case "replace":
		return fn("replace", str, posOnly(str), posOnly(str), optPosOnly(intT)), true
	case "split", "rsplit":
		return fn(name, listStr, optPosOnly(Union(str, None())), optPosOnly(intT)), true
	case "splitlines":
		return fn("splitlines", listStr, optPosOnly(boolT)), true
	case "partition", "rpartition":
		return fn(name, Tuple(str, str, str), posOnly(str)), true
	case "format":
		return fn("format", str, varArgs(Any()), varKwargs(Any())), true
	case "elems", "codepoints":
		return fn(name, Iter(str)), true
	case "elem_ords", "codepoint_ords":
		return fn(name, Iter(intT)), true
	}
	return Never(), false
}
