// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typecheck

// Module-level partial evaluation, the analogue of starlark-rust's
// fill_types_for_lint: a pre-pass over the file that maps each
// binding to the static value its assignments compute, when the
// evaluator understands it - a type value (IntList = list[int]), a
// string literal, or a value minted by a TypeFactory (record, enum,
// error_tags). This is what lets annotations refer to computed
// values: the checker distinguishes the *type of a binding*
// (IntList: type) from the *type value it denotes* (list[int]).
//
// Evaluation is flow-insensitive with agree-or-unknown merging: a
// binding assigned the same static value everywhere has that value;
// a disagreement (if/else branches assigning different types), an
// assignment the evaluator does not understand, or a non-assignment
// binding form (parameter, loop variable, catch variable) degrades
// the binding to unknown. Unknown bindings type as typing.Any plus an
// Approximation at the use site - the evaluator never guesses, and
// never rejects a program that runs.

import (
	"go.starlark.net/resolve"
	"go.starlark.net/syntax"
)

// A StaticValue is the partial evaluator's model of a value computed
// at module load time. The zero value means "not understood".
type StaticValue struct {
	denotes *Ty     // the type the value denotes in annotation position, if any
	bind    *Ty     // the precise type of the value itself, if known
	str     *string // string literal payload, if any
	data    any     // adapter-defined payload (see DataValue), if any
}

// TypeValue returns a static value that denotes the type t in
// annotation position, the value of a type alias like
// IntList = list[int].
func TypeValue(t Ty) StaticValue { return StaticValue{denotes: &t} }

// ValueOf returns a static value about which nothing is known but its
// type. TypeFactory implementations use it to describe minted values
// such as record constructors or error tag sets.
func ValueOf(ty Ty) StaticValue { return StaticValue{bind: &ty} }

// Denoting returns a copy of v that additionally denotes the type t
// in annotation position.
func (v StaticValue) Denoting(t Ty) StaticValue {
	v.denotes = &t
	return v
}

// DataValue returns a static value carrying an adapter-defined
// payload, retrievable with Data. The payload must be comparable
// (typically a pointer) so that assignments can be tested for
// agreement.
func DataValue(data any) StaticValue { return StaticValue{data: data} }

func strValue(s string) StaticValue { return StaticValue{str: &s} }

// DenotesType returns the type v denotes in annotation position.
func (v StaticValue) DenotesType() (Ty, bool) {
	if v.denotes == nil {
		return Never(), false
	}
	return *v.denotes, true
}

// Str returns v's string literal payload.
func (v StaticValue) Str() (string, bool) {
	if v.str == nil {
		return "", false
	}
	return *v.str, true
}

// Data returns v's adapter-defined payload.
func (v StaticValue) Data() (any, bool) {
	if v.data == nil {
		return nil, false
	}
	return v.data, true
}

func (v StaticValue) equal(w StaticValue) bool {
	eqTy := func(a, b *Ty) bool {
		if (a == nil) != (b == nil) {
			return false
		}
		return a == nil || a.Equal(*b)
	}
	eqStr := func(a, b *string) bool {
		if (a == nil) != (b == nil) {
			return false
		}
		return a == nil || *a == *b
	}
	return eqTy(v.denotes, w.denotes) && eqTy(v.bind, w.bind) &&
		eqStr(v.str, w.str) && v.data == w.data
}

// A TypeFactory interprets module-level calls of a builtin that
// creates a new type or typed value, such as record, enum, or
// error_tags. The partial evaluator invokes it when a call of the
// builtin appears as the right-hand side of an assignment and every
// argument was statically evaluated; the factory returns the static
// value of the call, typically minting a Custom type. Returning false
// means the call is not understood and the binding becomes unknown.
type TypeFactory func(call *FactoryCall) (StaticValue, bool)

// A FactoryCall describes a statically evaluated call of a builtin
// with an attached TypeFactory.
type FactoryCall struct {
	Name       string          // global to which the result is assigned, or ""
	Pos        syntax.Position // position of the call
	Positional []StaticValue
	Named      []NamedValue
}

// A NamedValue is a named argument of a FactoryCall.
type NamedValue struct {
	Name  string
	Value StaticValue
}

// WithFactory returns fn, which must be a single callable type (see
// Function), with f attached as its TypeFactory. The factory does not
// participate in type identity.
func WithFactory(fn Ty, f TypeFactory) Ty {
	if len(fn.alts) != 1 {
		panic("WithFactory: not a callable type")
	}
	cb, ok := fn.alts[0].(callableBasic)
	if !ok {
		panic("WithFactory: not a callable type")
	}
	cb.factory = f
	return Ty{alts: []Basic{cb}}
}

// A pevalEntry is the merged static value of one binding.
type pevalEntry struct {
	v        StaticValue
	conflict bool
}

func pevalKeyOf(id *syntax.Ident) (any, bool) {
	b, ok := id.Binding.(*resolve.Binding)
	if !ok || b == nil {
		return nil, false
	}
	return bindKeyOf(b), true
}

// pevalLookup returns the merged static value of id's binding.
func (c *checker) pevalLookup(id *syntax.Ident) (StaticValue, bool) {
	key, ok := pevalKeyOf(id)
	if !ok {
		return StaticValue{}, false
	}
	ent := c.peval[key]
	if ent == nil || ent.conflict {
		return StaticValue{}, false
	}
	return ent.v, true
}

func (c *checker) pevalMerge(id *syntax.Ident, v StaticValue) {
	key, ok := pevalKeyOf(id)
	if !ok {
		return
	}
	if ent := c.peval[key]; ent == nil {
		c.peval[key] = &pevalEntry{v: v}
	} else if !ent.conflict && !ent.v.equal(v) {
		ent.conflict = true // disagreement: unknown
	}
}

func (c *checker) pevalConflict(id *syntax.Ident) {
	key, ok := pevalKeyOf(id)
	if !ok {
		return
	}
	if ent := c.peval[key]; ent == nil {
		c.peval[key] = &pevalEntry{conflict: true}
	} else {
		ent.conflict = true
	}
}

// pevalModule runs the partial evaluation pre-pass. It must run
// before collection so that annotations anywhere in the file can
// refer to the computed values.
func (c *checker) pevalModule() {
	c.pevalStmts(c.file.Stmts)
}

func (c *checker) pevalStmts(stmts []syntax.Stmt) {
	for _, stmt := range stmts {
		c.pevalStmt(stmt)
	}
}

func (c *checker) pevalStmt(stmt syntax.Stmt) {
	switch stmt := stmt.(type) {
	case *syntax.AssignStmt:
		if stmt.Op == syntax.EQ {
			if id, ok := unparen(stmt.LHS).(*syntax.Ident); ok {
				name := ""
				if b, _ := id.Binding.(*resolve.Binding); b != nil && b.Scope == resolve.Global {
					name = id.Name // like the runtime's Exportable
				}
				if v, ok := c.staticEval(stmt.RHS, name); ok {
					c.pevalMerge(id, v)
				} else {
					c.pevalConflict(id)
				}
				c.pevalExprBindings(stmt.RHS)
				return
			}
		}
		c.pevalConflictTargets(stmt.LHS)
		c.pevalExprBindings(stmt.RHS)

	case *syntax.DefStmt:
		c.pevalConflict(stmt.Name)
		for _, param := range stmt.Params {
			c.pevalConflictParam(param)
		}
		c.pevalStmts(stmt.Body)

	case *syntax.ForStmt:
		c.pevalConflictTargets(stmt.Vars)
		c.pevalExprBindings(stmt.X)
		c.pevalStmts(stmt.Body)

	case *syntax.WhileStmt:
		c.pevalExprBindings(stmt.Cond)
		c.pevalStmts(stmt.Body)

	case *syntax.IfStmt:
		c.pevalExprBindings(stmt.Cond)
		c.pevalStmts(stmt.True)
		c.pevalStmts(stmt.False)

	case *syntax.LoadStmt:
		iface := c.loads[stmt.ModuleName()]
		for i, to := range stmt.To {
			if t, ok := iface.Denoted(stmt.From[i].Name); ok {
				c.pevalMerge(to, TypeValue(t))
			} else {
				c.pevalConflict(to)
			}
		}

	case *syntax.ExprStmt:
		c.pevalExprBindings(stmt.X)

	case *syntax.ReturnStmt:
		c.pevalExprBindings(stmt.Result)

	case *syntax.DeferStmt:
		c.pevalExprBindings(stmt.Call)

	case *syntax.ErrDeferStmt:
		c.pevalExprBindings(stmt.Call)

	case *syntax.RecoverStmt:
		c.pevalExprBindings(stmt.Result)
	}
}

// pevalConflictParam degrades a function parameter's binding.
func (c *checker) pevalConflictParam(param syntax.Expr) {
	switch param := param.(type) {
	case *syntax.Ident:
		c.pevalConflict(param)
	case *syntax.BinaryExpr: // name=default
		if id, ok := param.X.(*syntax.Ident); ok {
			c.pevalConflict(id)
		}
		c.pevalExprBindings(param.Y)
	case *syntax.UnaryExpr: // *args, **kwargs, bare * or /
		if id, ok := param.X.(*syntax.Ident); ok {
			c.pevalConflict(id)
		}
	case *syntax.TypedParam:
		c.pevalConflictParam(param.X)
		c.pevalExprBindings(param.Default)
	}
}

// pevalConflictTargets degrades the bindings of an assignment target.
func (c *checker) pevalConflictTargets(lhs syntax.Expr) {
	switch lhs := unparen(lhs).(type) {
	case *syntax.Ident:
		c.pevalConflict(lhs)
	case *syntax.TupleExpr:
		for _, elem := range lhs.List {
			c.pevalConflictTargets(elem)
		}
	case *syntax.ListExpr:
		for _, elem := range lhs.List {
			c.pevalConflictTargets(elem)
		}
	case *syntax.IndexExpr:
		c.pevalExprBindings(lhs.X)
		c.pevalExprBindings(lhs.Y)
	case *syntax.DotExpr:
		c.pevalExprBindings(lhs.X)
	}
}

// pevalExprBindings degrades bindings created inside an expression:
// comprehension and lambda variables, catch error variables, and any
// assignment inside a catch block.
func (c *checker) pevalExprBindings(e syntax.Expr) {
	if e == nil {
		return
	}
	syntax.Walk(e, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.ForClause:
			c.pevalConflictTargets(n.Vars)
		case *syntax.CatchExpr:
			if n.ErrorVar != nil {
				c.pevalConflict(n.ErrorVar)
			}
		case *syntax.LambdaExpr:
			for _, param := range n.Params {
				c.pevalConflictParam(param)
			}
		case *syntax.AssignStmt: // inside a catch block
			c.pevalConflictTargets(n.LHS)
		}
		return true
	})
}

// staticEval computes the static value of expression e, if the
// evaluator understands it. name is the global to which the value is
// being assigned, or ""; it names types minted by factories, like the
// runtime's Exportable.
func (c *checker) staticEval(e syntax.Expr, name string) (StaticValue, bool) {
	if t, ok := c.strictTypeExpr(e); ok {
		return TypeValue(t), true
	}
	switch e := unparen(e).(type) {
	case *syntax.Literal:
		switch e.Token {
		case syntax.STRING:
			return strValue(e.Value.(string)), true
		case syntax.INT:
			return ValueOf(Prim("int")), true
		case syntax.FLOAT:
			return ValueOf(Prim("float")), true
		case syntax.BYTES:
			return ValueOf(Prim("bytes")), true
		}
	case *syntax.Ident:
		if b, _ := e.Binding.(*resolve.Binding); b != nil && b.Scope == resolve.Universal {
			switch e.Name {
			case "True", "False":
				return ValueOf(Prim("bool")), true
			}
			return StaticValue{}, false
		}
		return c.pevalLookup(e)
	case *syntax.CondExpr:
		x, okx := c.staticEval(e.True, name)
		y, oky := c.staticEval(e.False, name)
		if okx && oky && x.equal(y) {
			return x, true
		}
	case *syntax.CallExpr:
		return c.staticEvalCall(e, name)
	}
	return StaticValue{}, false
}

// strictTypeExpr interprets e as a type expression, failing (rather
// than approximating) on anything unknown.
func (c *checker) strictTypeExpr(e syntax.Expr) (Ty, bool) {
	before := len(c.o.approx)
	t := c.o.tyFromTypeExpr(e, c.lookupAnnot)
	if len(c.o.approx) == before {
		return t, true
	}
	c.o.approx = c.o.approx[:before]
	return Never(), false
}

// staticEvalCall evaluates a call whose callee names a universal or
// predeclared builtin with an attached TypeFactory.
func (c *checker) staticEvalCall(call *syntax.CallExpr, name string) (StaticValue, bool) {
	id, ok := unparen(call.Fn).(*syntax.Ident)
	if !ok {
		return StaticValue{}, false
	}
	b, ok := id.Binding.(*resolve.Binding)
	if !ok || b == nil {
		return StaticValue{}, false
	}
	var fnTy Ty
	switch b.Scope {
	case resolve.Universal:
		fnTy = universeTypes[id.Name]
	case resolve.Predeclared:
		fnTy = c.env[id.Name]
	default:
		return StaticValue{}, false
	}
	var factory TypeFactory
	if len(fnTy.alts) == 1 {
		if cb, ok := fnTy.alts[0].(callableBasic); ok {
			factory = cb.factory
		}
	}
	if factory == nil {
		return StaticValue{}, false
	}
	fc := &FactoryCall{Name: name, Pos: exprPos(call)}
	for _, arg := range call.Args {
		switch arg := arg.(type) {
		case *syntax.BinaryExpr:
			if arg.Op == syntax.EQ {
				if argName, ok := arg.X.(*syntax.Ident); ok {
					v, ok := c.staticEval(arg.Y, "")
					if !ok {
						return c.factoryUnknown(id.Name, arg.Y)
					}
					fc.Named = append(fc.Named, NamedValue{argName.Name, v})
					continue
				}
			}
		case *syntax.UnaryExpr:
			if arg.Op == syntax.STAR || arg.Op == syntax.STARSTAR {
				return c.factoryUnknown(id.Name, arg)
			}
		}
		v, ok := c.staticEval(arg, "")
		if !ok {
			return c.factoryUnknown(id.Name, arg)
		}
		fc.Positional = append(fc.Positional, v)
	}
	return factory(fc)
}

// factoryUnknown records that a factory call could not be evaluated.
func (c *checker) factoryUnknown(fn string, arg syntax.Expr) (StaticValue, bool) {
	c.o.approximation("Partial evaluation",
		"call of %s at %s with an argument the evaluator does not understand", fn, exprPos(arg))
	return StaticValue{}, false
}
