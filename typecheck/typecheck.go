// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package typecheck provides optional static type analysis for
// Starlark files, a port of the typechecker of starlark-rust
// (https://github.com/facebook/starlark-rust, starlark/src/typing).
//
// The checker runs on a resolved syntax tree and is entirely separate
// from execution: nothing in the starlark interpreter calls it. It is
// deliberately lenient ("approximation-based"): when it cannot model
// a construct it assumes typing.Any, records an Approximation, and
// moves on. It aims never to report an error for a program that would
// execute successfully.
package typecheck // import "go.starlark.net/typecheck"

import (
	"fmt"
	"maps"
	"sort"
	"strings"

	"go.starlark.net/resolve"
	"go.starlark.net/syntax"
)

// An Env maps predeclared and universal names to their types.
// Names absent from the Env type as Any.
type Env map[string]Ty

// An Error is a static type error.
type Error struct {
	Pos syntax.Position
	Msg string
}

func (e Error) Error() string { return fmt.Sprintf("%s: %s", e.Pos, e.Msg) }

// An Approximation records a place where the checker could not model
// the program precisely and assumed typing.Any instead.
type Approximation struct {
	Category string
	Message  string
}

func (a Approximation) String() string { return a.Category + ": " + a.Message }

// A Result holds the outcome of a Check.
type Result struct {
	Errors         []Error         // type errors (analysis continues past them)
	Types          *TypeMap        // inferred type of each binding
	Interface      *Interface      // types of the module's globals
	Approximations []Approximation // where the checker punted
}

// A TypeMap records the inferred type of every binding in a file.
type TypeMap struct {
	m map[any]bindingInfo // keyed by binding identity (see bindKey)
}

type bindingInfo struct {
	name string
	pos  syntax.Position
	ty   Ty
}

// Lookup returns the inferred type of the binding of id, which must
// be an identifier of the checked file.
func (tm *TypeMap) Lookup(id *syntax.Ident) (Ty, bool) {
	b, ok := id.Binding.(*resolve.Binding)
	if !ok || b == nil {
		return Never(), false
	}
	info, ok := tm.m[bindKeyOf(b)]
	return info.ty, ok
}

// String formats the map as sorted "name (pos) = type" lines,
// for tests and debugging.
func (tm *TypeMap) String() string {
	var lines []string
	for _, info := range tm.m {
		if info.name == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s (%s) = %s", info.name, info.pos, info.ty))
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

// An Interface describes the types of a module's exported globals,
// for use as the loads argument when checking dependent files.
type Interface struct {
	types map[string]Ty
}

// NewInterface returns an Interface with the given member types.
func NewInterface(types map[string]Ty) *Interface {
	m := make(map[string]Ty, len(types))
	maps.Copy(m, types)
	return &Interface{types: m}
}

// Get returns the type of the named module member.
func (i *Interface) Get(name string) (Ty, bool) {
	if i == nil {
		return Never(), false
	}
	ty, ok := i.types[name]
	return ty, ok
}

// Names returns the sorted member names.
func (i *Interface) Names() []string {
	names := make([]string, 0, len(i.types))
	for name := range i.types {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// bindKeyOf returns the canonical identity of a binding. Free-variable
// bindings are created fresh for each capturing function but share
// their First identifier with the original local binding, so keying by
// First unifies them.
func bindKeyOf(b *resolve.Binding) any {
	if b.First != nil {
		return b.First
	}
	return b
}

// Check performs static type analysis on a resolved file.
//
// The file must have been resolved (e.g. by starlark.ExecFile, or
// resolve.File) so that identifiers carry binding information; Check
// returns an error otherwise. env provides the types of predeclared
// and universal names (see UniverseEnv); loads provides the
// Interfaces of modules the file loads, keyed by module name.
//
// Type errors are reported in Result.Errors; the error result is
// reserved for usage errors.
func Check(file *syntax.File, env Env, loads map[string]*Interface) (*Result, error) {
	if file.Module == nil {
		return nil, fmt.Errorf("typecheck: file %q has not been resolved", file.Path)
	}

	c := newChecker(file, env, loads)
	c.collectAliases()
	c.collectStmts(file.Stmts)
	c.solve()
	c.check()

	// Build the Interface from the module's globals.
	module := file.Module.(*resolve.Module)
	ifaceTypes := make(map[string]Ty, len(module.Globals))
	for _, g := range module.Globals {
		name := "?"
		if g.First != nil {
			name = g.First.Name
		}
		ifaceTypes[name] = c.typeOfBinding(g)
	}

	// Sort errors by position for deterministic output.
	sort.SliceStable(c.o.errors, func(i, j int) bool {
		ei, ej := c.o.errors[i].Pos, c.o.errors[j].Pos
		if ei.Line != ej.Line {
			return ei.Line < ej.Line
		}
		return ei.Col < ej.Col
	})

	return &Result{
		Errors:         c.o.errors,
		Types:          c.typeMap(),
		Interface:      NewInterface(ifaceTypes),
		Approximations: c.o.approx,
	}, nil
}
