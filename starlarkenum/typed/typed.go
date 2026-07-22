// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package typed contributes a static type for the enum builtin of
// go.starlark.net/starlarkenum to a typecheck environment, so that
// the static checker understands module-level enum definitions:
//
//	Color = enum("red", "green", "blue")
//	c = Color("red")
//
// The checker then reports unknown elements (Color.bleu) and
// constructor argument errors, and the annotation `x: Color` matches
// exactly the elements of Color, like the runtime matcher. Identity
// is nominal: two structurally identical enum types do not match each
// other's elements.
//
// Use it alongside starlarkenum.Predeclared:
//
//	env := typecheck.UniverseEnv()
//	typed.AddTypes(env)
package typed // import "go.starlark.net/starlarkenum/typed"

import (
	"go.starlark.net/typecheck"
)

// AddTypes adds the static type of the enum builtin to env. Its
// runtime counterpart is starlarkenum.Predeclared.
func AddTypes(env typecheck.Env) {
	env["enum"] = typecheck.WithFactory(
		typecheck.Function("enum",
			&typecheck.ParamSpec{Params: []typecheck.Param{
				{Mode: typecheck.ArgsMode, Ty: typecheck.Prim("string")},
			}},
			typecheck.Any()),
		enumFactory)
}

// An enumValueTy is the static type of the elements of one enum type.
type enumValueTy struct {
	name string
}

func (e *enumValueTy) TyName() string { return e.name }

func (e *enumValueTy) Attr(name string) (typecheck.Ty, bool) {
	switch name {
	case "value":
		return typecheck.Prim("string"), true
	case "index":
		return typecheck.Prim("int"), true
	}
	return typecheck.Never(), false
}

// An enumTy is the static type of the enum type value itself: a
// callable (constructing elements), indexable and iterable (over
// elements), with one attribute per element plus values().
type enumTy struct {
	name     string
	instance typecheck.Ty // Custom of the enumValueTy
	elements []string
}

func (e *enumTy) TyName() string { return "type[" + e.name + "]" }

func (e *enumTy) Attr(name string) (typecheck.Ty, bool) {
	if name == "values" {
		return typecheck.Function("values",
			typecheck.PositionalOnly(),
			typecheck.List(typecheck.Prim("string"))), true
	}
	for _, el := range e.elements {
		if el == name {
			return e.instance, true
		}
	}
	return typecheck.Never(), false
}

func (e *enumTy) CallSignature() (*typecheck.ParamSpec, typecheck.Ty) {
	// The runtime accepts an element string or an existing element.
	arg := typecheck.Union(typecheck.Prim("string"), e.instance)
	return typecheck.PositionalOnly(arg), e.instance
}

func (e *enumTy) IndexResult(index typecheck.Ty) (typecheck.Ty, bool) {
	if typecheck.Intersects(index, typecheck.Prim("int")) {
		return e.instance, true
	}
	return typecheck.Never(), false
}

func (e *enumTy) IterItem() typecheck.Ty { return e.instance }

// enumFactory interprets Color = enum("red", "green", "blue"). It
// needs every element to be a string literal; anything else degrades
// the binding to unknown.
func enumFactory(call *typecheck.FactoryCall) (typecheck.StaticValue, bool) {
	if len(call.Named) > 0 || len(call.Positional) == 0 {
		return typecheck.StaticValue{}, false
	}
	name := call.Name
	if name == "" {
		name = "enum" // like the runtime's displayName fallback
	}
	elements := make([]string, len(call.Positional))
	for i, arg := range call.Positional {
		s, ok := arg.Str()
		if !ok {
			return typecheck.StaticValue{}, false
		}
		elements[i] = s
	}
	instance := typecheck.Custom(&enumValueTy{name: name})
	et := &enumTy{name: name, instance: instance, elements: elements}
	return typecheck.ValueOf(typecheck.Custom(et)).Denoting(instance), true
}
