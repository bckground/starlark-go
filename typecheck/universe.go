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

import (
	"maps"

	"go.starlark.net/syntax"
)

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

// fnS is fn with a specialFunc that refines the result type from the
// argument types; result remains the (sound) fallback.
func fnS(name string, special specialFunc, result Ty, params ...Param) Ty {
	return Ty{alts: []Basic{callableBasic{
		name:      name,
		params:    &ParamSpec{Params: params},
		result:    result,
		specialFn: special,
	}}}
}

// Result-refining specials. Each bails out (false) on *args/**kwargs
// call forms, whose shapes are unknown; the declared result type, a
// superset, then applies. Leniency requires refinements to be sound
// supersets of the runtime result, never wishful narrowings.

// plainArgs reports whether the call uses neither *args nor **kwargs.
func plainArgs(args callArgs) bool {
	return args.starArgs == nil && args.kwargsArgs == nil
}

// absSpecial: abs(int) is int, abs(float) is float.
func absSpecial(o *oracle, pos syntax.Position, args callArgs) (Ty, bool) {
	if !plainArgs(args) || len(args.pos) != 1 {
		return Never(), false
	}
	if t := args.pos[0].ty; t.Equal(Prim("int")) || t.Equal(Prim("float")) {
		return t, true
	}
	return Never(), false
}

// minMaxSpecial: min/max of one iterable yields its element type; of
// several arguments, their union.
func minMaxSpecial(o *oracle, pos syntax.Position, args callArgs) (Ty, bool) {
	if !plainArgs(args) {
		return Never(), false
	}
	switch len(args.pos) {
	case 0:
		return Never(), false
	case 1:
		return o.iterItem(args.pos[0].pos, args.pos[0].ty), true
	default:
		tys := make([]Ty, len(args.pos))
		for i, a := range args.pos {
			tys[i] = a.ty
		}
		return Union(tys...), true
	}
}

// elemListSpecial: sorted/reversed yield a list of the argument's
// element type.
func elemListSpecial(o *oracle, pos syntax.Position, args callArgs) (Ty, bool) {
	if !plainArgs(args) || len(args.pos) < 1 {
		return Never(), false
	}
	return List(o.iterItem(args.pos[0].pos, args.pos[0].ty)), true
}

// listSpecial: list() is empty; list(x) holds x's elements.
func listSpecial(o *oracle, pos syntax.Position, args callArgs) (Ty, bool) {
	if !plainArgs(args) {
		return Never(), false
	}
	if len(args.pos) == 0 {
		return List(Never()), true
	}
	return List(o.iterItem(args.pos[0].pos, args.pos[0].ty)), true
}

// setSpecial: like listSpecial, for set.
func setSpecial(o *oracle, pos syntax.Position, args callArgs) (Ty, bool) {
	if !plainArgs(args) {
		return Never(), false
	}
	if len(args.pos) == 0 {
		return Set(Never()), true
	}
	return Set(o.iterItem(args.pos[0].pos, args.pos[0].ty)), true
}

// tupleSpecial: tuple() is the empty tuple; tuple(x) has unknown arity
// over x's element type.
func tupleSpecial(o *oracle, pos syntax.Position, args callArgs) (Ty, bool) {
	if !plainArgs(args) {
		return Never(), false
	}
	if len(args.pos) == 0 {
		return Tuple(), true
	}
	return TupleOf(o.iterItem(args.pos[0].pos, args.pos[0].ty)), true
}

// zipSpecial: zip is arity-aware — zip(xs, ys) yields
// list[tuple[elem(xs), elem(ys)]].
func zipSpecial(o *oracle, pos syntax.Position, args callArgs) (Ty, bool) {
	if !plainArgs(args) {
		return Never(), false
	}
	elems := make([]Ty, len(args.pos))
	for i, a := range args.pos {
		elems[i] = o.iterItem(a.pos, a.ty)
	}
	return List(Tuple(elems...)), true
}

// enumerateSpecial: enumerate(xs) yields list[tuple[int, elem(xs)]].
func enumerateSpecial(o *oracle, pos syntax.Position, args callArgs) (Ty, bool) {
	if !plainArgs(args) || len(args.pos) < 1 {
		return Never(), false
	}
	return List(Tuple(Prim("int"), o.iterItem(args.pos[0].pos, args.pos[0].ty))), true
}

// dictSpecial: dict() from a dict argument, an iterable of pairs, or
// keyword arguments.
func dictSpecial(o *oracle, pos syntax.Position, args callArgs) (Ty, bool) {
	if !plainArgs(args) || len(args.pos) > 1 {
		return Never(), false
	}
	key, val := Never(), Never()
	if len(args.pos) == 1 {
		t := args.pos[0].ty
		if t.IsAny() {
			return Never(), false
		}
		for _, alt := range t.alts {
			if d, ok := alt.(dictBasic); ok {
				key, val = Union(key, d.key), Union(val, d.val)
				continue
			}
			// An iterable of fixed pairs: dict([(k, v), ...]).
			item, ok := iterItemBasic(alt)
			if !ok {
				return Never(), false
			}
			k, v, ok := pairElems(item)
			if !ok {
				return Never(), false
			}
			key, val = Union(key, k), Union(val, v)
		}
	}
	if len(args.named) > 0 {
		key = Union(key, Prim("string"))
		for _, na := range args.named {
			val = Union(val, na.ty)
		}
	}
	return Dict(key, val), true
}

// pairElems decomposes t into the components of a two-element tuple,
// if every alternative of t is one.
func pairElems(t Ty) (k, v Ty, ok bool) {
	if t.IsAny() || t.IsNever() {
		return Never(), Never(), false
	}
	k, v = Never(), Never()
	for _, alt := range t.alts {
		tup, isTuple := alt.(tupleBasic)
		if !isTuple || tup.of != nil || len(tup.elems) != 2 {
			return Never(), Never(), false
		}
		k, v = Union(k, tup.elems[0]), Union(v, tup.elems[1])
	}
	return k, v, true
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

		"abs":       fnS("abs", absSpecial, num, posOnly(num)),
		"any":       fn("any", boolT, posOnly(iterAny)),
		"all":       fn("all", boolT, posOnly(iterAny)),
		"bool":      fn("bool", boolT, optPosOnly(Any())),
		"bytes":     fn("bytes", Prim("bytes"), posOnly(Any())),
		"chr":       fn("chr", str, posOnly(intT)),
		"dict":      fnS("dict", dictSpecial, Dict(Any(), Any()), optPosOnly(Any()), varKwargs(Any())),
		"dir":       fn("dir", List(str), posOnly(Any())),
		"enumerate": fnS("enumerate", enumerateSpecial, List(Tuple(intT, Any())), posOnly(iterAny), optPosOnly(intT)),
		"fail":      fn("fail", Never(), varArgs(Any()), nameOnly("sep", str)),
		"float":     fn("float", Prim("float"), optPosOnly(Any())),
		"getattr":   fn("getattr", Any(), posOnly(Any()), posOnly(str), optPosOnly(Any())),
		"hasattr":   fn("hasattr", boolT, posOnly(Any()), posOnly(str)),
		"hash":      fn("hash", intT, posOnly(Union(str, Prim("bytes")))),
		"int":       fn("int", intT, optPosOnly(Any()), opt("base", intT)),
		"len":       fn("len", intT, posOnly(Any())),
		"list":      fnS("list", listSpecial, List(Any()), optPosOnly(iterAny)),
		"max":       fnS("max", minMaxSpecial, Any(), varArgs(Any()), nameOnly("key", Any())),
		"min":       fnS("min", minMaxSpecial, Any(), varArgs(Any()), nameOnly("key", Any())),
		"ord":       fn("ord", intT, posOnly(Union(str, Prim("bytes")))),
		"print":     fn("print", None(), varArgs(Any()), nameOnly("sep", str)),
		"range":     fn("range", Prim("range"), posOnly(intT), optPosOnly(intT), optPosOnly(intT)),
		"repr":      fn("repr", str, posOnly(Any())),
		"reversed":  fnS("reversed", elemListSpecial, List(Any()), posOnly(iterAny)),
		"set":       fnS("set", setSpecial, Set(Any()), optPosOnly(iterAny)),
		"sorted":    fnS("sorted", elemListSpecial, List(Any()), posOnly(iterAny), nameOnly("key", Any()), nameOnly("reverse", boolT)),
		"str":       fn("str", str, optPosOnly(Any())),
		"tuple":     fnS("tuple", tupleSpecial, AnyTuple(), optPosOnly(iterAny)),
		"type":      fn("type", str, posOnly(Any())),
		"zip":       fnS("zip", zipSpecial, List(AnyTuple()), varArgs(iterAny)),

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
	// get(k) -> V | None, get(k, d) -> V | typeof(d), and similarly
	// pop and setdefault: the default argument's type joins the
	// result, so the precise result is computed per call site.
	withDefault := func(oneArg Ty) specialFunc {
		return func(o *oracle, pos syntax.Position, args callArgs) (Ty, bool) {
			if !plainArgs(args) {
				return Never(), false
			}
			switch len(args.pos) {
			case 1:
				return oneArg, true
			case 2:
				return Union(val, args.pos[1].ty), true
			}
			return Never(), false
		}
	}
	switch name {
	case "clear":
		return fn("clear", None()), true
	case "get":
		return fnS("get", withDefault(Union(val, None())), Any(), posOnly(key), optPosOnly(Any())), true
	case "items":
		return fn("items", List(Tuple(key, val))), true
	case "keys":
		return fn("keys", List(key)), true
	case "pop":
		return fnS("pop", withDefault(val), Any(), posOnly(key), optPosOnly(Any())), true
	case "popitem":
		return fn("popitem", Tuple(key, val)), true
	case "setdefault":
		return fnS("setdefault", withDefault(Union(val, None())), Any(), posOnly(key), optPosOnly(Any())), true
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
