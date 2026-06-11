// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typecheck

// Static interpretation of type-annotation expressions: the analogue
// of the runtime's starlark.TypeOf. The two interpretations must
// agree; see TestAnnotationAgreement.

import (
	"go.starlark.net/syntax"
)

// tyLookup resolves a name used in a type-annotation expression to a
// type. The bool result reports whether the name was recognized.
type tyLookup func(name string) (Ty, bool)

// builtinTypeName returns the type denoted by a canonical type
// constructor name in annotation position.
func builtinTypeName(name string) (Ty, bool) {
	switch name {
	case "int", "bool", "float", "bytes", "range":
		return Prim(name), true
	case "str":
		return Prim("string"), true
	case "None":
		return None(), true
	case "list":
		return List(Any()), true
	case "dict":
		return Dict(Any(), Any()), true
	case "tuple":
		return AnyTuple(), true
	case "set":
		return Set(Any()), true
	case "type":
		return TypeType(), true
	}
	return Never(), false
}

// tyFromTypeExpr converts a type-annotation expression to a Ty.
// Unknown constructs yield Any plus an Approximation: the static
// checker must never reject an annotation the runtime accepts.
func (o *oracle) tyFromTypeExpr(e syntax.Expr, lookup tyLookup) Ty {
	unknown := func(what string) Ty {
		o.approximation("Unknown type", "%s in type annotation", what)
		return Any()
	}
	switch e := e.(type) {
	case *syntax.Ident:
		if t, ok := builtinTypeName(e.Name); ok {
			return t
		}
		if lookup != nil {
			if t, ok := lookup(e.Name); ok {
				return t
			}
		}
		return unknown("name `" + e.Name + "`")

	case *syntax.ParenExpr:
		return o.tyFromTypeExpr(e.X, lookup)

	case *syntax.DotExpr:
		// typing.Any etc.
		if x, ok := e.X.(*syntax.Ident); ok && x.Name == "typing" {
			switch e.Name.Name {
			case "Any":
				return Any()
			case "Never":
				return Never()
			case "Callable":
				return AnyCallable()
			case "Iterable":
				return Iter(Any())
			}
		}
		return unknown("dotted path")

	case *syntax.BinaryExpr:
		if e.Op == syntax.PIPE {
			return Union(o.tyFromTypeExpr(e.X, lookup), o.tyFromTypeExpr(e.Y, lookup))
		}
		return unknown("operator")

	case *syntax.TupleExpr:
		elems := make([]Ty, len(e.List))
		for i, x := range e.List {
			elems[i] = o.tyFromTypeExpr(x, lookup)
		}
		return Tuple(elems...)

	case *syntax.ListExpr:
		// legacy union syntax [T1, T2]
		tys := make([]Ty, len(e.List))
		for i, x := range e.List {
			tys[i] = o.tyFromTypeExpr(x, lookup)
		}
		return Union(tys...)

	case *syntax.IndexExpr:
		return o.tyFromIndexExpr(e, lookup)
	}
	return unknown("expression")
}

func (o *oracle) tyFromIndexExpr(e *syntax.IndexExpr, lookup tyLookup) Ty {
	args := []syntax.Expr{e.Y}
	if tuple, ok := e.Y.(*syntax.TupleExpr); ok {
		args = tuple.List
	}
	arg := func(i int) Ty { return o.tyFromTypeExpr(args[i], lookup) }

	unknown := func(what string) Ty {
		o.approximation("Unknown type", "%s in type annotation", what)
		return Any()
	}

	switch x := e.X.(type) {
	case *syntax.Ident:
		switch x.Name {
		case "list":
			if len(args) == 1 {
				return List(arg(0))
			}
		case "set":
			if len(args) == 1 {
				return Set(arg(0))
			}
		case "dict":
			if len(args) == 2 {
				return Dict(arg(0), arg(1))
			}
		case "tuple":
			// tuple[T, ...] or tuple[T1, T2]
			if len(args) == 2 {
				if _, ok := args[1].(*syntax.EllipsisExpr); ok {
					return TupleOf(arg(0))
				}
			}
			elems := make([]Ty, len(args))
			for i := range args {
				if _, ok := args[i].(*syntax.EllipsisExpr); ok {
					return unknown("misplaced `...`")
				}
				elems[i] = arg(i)
			}
			return Tuple(elems...)
		}
		return unknown("parameterized name `" + x.Name + "`")

	case *syntax.DotExpr:
		if xx, ok := x.X.(*syntax.Ident); ok && xx.Name == "typing" {
			switch x.Name.Name {
			case "Callable":
				// typing.Callable[[T1, T2], R] or typing.Callable[..., R]
				if len(args) == 2 {
					result := arg(1)
					switch params := args[0].(type) {
					case *syntax.EllipsisExpr:
						return Callable(AnyParams(), result)
					case *syntax.ListExpr:
						types := make([]Ty, len(params.List))
						for i, p := range params.List {
							types[i] = o.tyFromTypeExpr(p, lookup)
						}
						return Callable(PositionalOnly(types...), result)
					}
				}
			case "Iterable":
				if len(args) == 1 {
					return Iter(arg(0))
				}
			}
		}
	}
	return unknown("parameterized expression")
}
