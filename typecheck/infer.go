// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typecheck

// The inference engine: bindings collection over the resolved syntax
// tree, a naive fixpoint solver, and per-expression type inference.
// It is the Go analogue of starlark-rust's BindingsCollect
// (typing/bindings.rs) and TypingContext (typing/ctx.rs).
//
// Each binding (as identified by bindKeyOf) accumulates a set of
// "binding expressions" describing the values assigned to it. The
// solver starts every binding at its declared type (or Never) and
// repeatedly unions in the types of its binding expressions until
// nothing changes.

import (
	"go.starlark.net/resolve"
	"go.starlark.net/syntax"
)

type checker struct {
	o     *oracle
	file  *syntax.File
	env   Env
	loads map[string]*Interface

	aliasesByName map[string]Ty // module-level type aliases

	declared map[any]Ty
	exprs    map[any][]bindExpr
	order    []any // deterministic solve order of binding keys
	infos    map[any]bindingInfo

	checks     []syntax.Expr
	checkTypes []checkTypeItem

	types map[any]Ty // current solution

	retStack   []*Ty                               // declared return types of enclosing defs (nil if unannotated)
	catchStack []*syntax.CatchExpr                 // enclosing catch blocks
	catchTypes map[*syntax.CatchExpr][]syntax.Expr // recover result exprs per catch block
}

type checkTypeItem struct {
	pos  syntax.Position
	expr syntax.Expr // nil for a bare `return`
	want Ty
}

func newChecker(file *syntax.File, env Env, loads map[string]*Interface) *checker {
	return &checker{
		o:             &oracle{typeMap: &TypeMap{m: make(map[any]bindingInfo)}},
		file:          file,
		env:           env,
		loads:         loads,
		aliasesByName: make(map[string]Ty),
		declared:      make(map[any]Ty),
		exprs:         make(map[any][]bindExpr),
		infos:         make(map[any]bindingInfo),
		types:         make(map[any]Ty),
		catchTypes:    make(map[*syntax.CatchExpr][]syntax.Expr),
	}
}

func exprPos(e syntax.Expr) syntax.Position {
	pos, _ := e.Span()
	return pos
}

// Binding expressions.

type bindExpr interface {
	// ty returns this binding site's contribution under the current
	// solution.
	ty(c *checker) Ty
}

// x = e
type exprBind struct{ e syntax.Expr }

func (b exprBind) ty(c *checker) Ty { return c.exprType(b.e) }

// for x in <of>
type iterBind struct {
	of  bindExpr
	pos syntax.Position
}

func (b iterBind) ty(c *checker) Ty { return c.o.iterItem(b.pos, b.of.ty(c)) }

// x, y = <of> (component i)
type indexBind struct {
	i  int
	of bindExpr
}

func (b indexBind) ty(c *checker) Ty {
	t := b.of.ty(c)
	if t.IsAny() || t.IsNever() {
		return t
	}
	var results []Ty
	for _, alt := range t.alts {
		switch alt := alt.(type) {
		case tupleBasic:
			if alt.of != nil {
				results = append(results, *alt.of)
			} else if b.i < len(alt.elems) {
				results = append(results, alt.elems[b.i])
			}
		case listBasic:
			results = append(results, alt.elem)
		default:
			if item, ok := iterItemBasic(alt); ok {
				results = append(results, item)
			} else {
				results = append(results, Any())
			}
		}
	}
	return Union(results...)
}

// x op= e
type modifyBind struct {
	lhs *syntax.Ident
	op  syntax.Token // PLUS, MINUS, ... (already de-augmented)
	rhs syntax.Expr
	pos syntax.Position
}

func (b modifyBind) ty(c *checker) Ty {
	return c.o.binop(b.pos, b.op, c.exprType(b.lhs), c.exprType(b.rhs))
}

// x.append(e): contributes list[typeof(e)]
type appendBind struct{ e syntax.Expr }

func (b appendBind) ty(c *checker) Ty { return List(c.exprType(b.e)) }

// x.extend(e): contributes list[iterItem(typeof(e))]
type extendBind struct {
	e   syntax.Expr
	pos syntax.Position
}

func (b extendBind) ty(c *checker) Ty {
	return List(c.o.iterItem(b.pos, c.exprType(b.e)))
}

// x[i] = v: refines container element types
type setIndexBind struct {
	target *syntax.Ident
	index  syntax.Expr
	val    syntax.Expr
}

func (b setIndexBind) ty(c *checker) Ty {
	t := c.exprType(b.target)
	idx := c.exprType(b.index)
	val := c.exprType(b.val)
	var results []Ty
	for _, alt := range t.alts {
		switch alt := alt.(type) {
		case listBasic:
			results = append(results, List(Union(alt.elem, val)))
		case dictBasic:
			results = append(results, Dict(Union(alt.key, idx), Union(alt.val, val)))
		}
	}
	return Union(results...) // Never if not a container (yet)
}

// a fixed type contribution
type tyBind struct{ t Ty }

func (b tyBind) ty(c *checker) Ty { return b.t }

// Collection.

func (c *checker) bindingOf(id *syntax.Ident) (any, bool) {
	b, ok := id.Binding.(*resolve.Binding)
	if !ok || b == nil {
		return nil, false
	}
	key := bindKeyOf(b)
	if _, seen := c.infos[key]; !seen {
		name, pos := id.Name, id.NamePos
		if b.First != nil {
			name, pos = b.First.Name, b.First.NamePos
		}
		c.infos[key] = bindingInfo{name: name, pos: pos}
		c.order = append(c.order, key)
	}
	return key, true
}

func (c *checker) declare(id *syntax.Ident, ty Ty) {
	if key, ok := c.bindingOf(id); ok {
		if old, exists := c.declared[key]; exists {
			c.declared[key] = Union(old, ty)
		} else {
			c.declared[key] = ty
		}
	}
}

func (c *checker) bind(id *syntax.Ident, be bindExpr) {
	if key, ok := c.bindingOf(id); ok {
		c.exprs[key] = append(c.exprs[key], be)
	}
}

// collectAliases records module-level type aliases (IntList = list[int])
// so that later annotations can refer to them by name.
func (c *checker) collectAliases() {
	for _, stmt := range c.file.Stmts {
		assign, ok := stmt.(*syntax.AssignStmt)
		if !ok || assign.Op != syntax.EQ {
			continue
		}
		id, ok := assign.LHS.(*syntax.Ident)
		if !ok {
			continue
		}
		if ty, ok := c.aliasTy(assign.RHS); ok {
			c.aliasesByName[id.Name] = ty
		}
	}
}

// aliasTy interprets e as a type expression, failing (rather than
// approximating) on anything unknown. Used only for alias detection.
func (c *checker) aliasTy(e syntax.Expr) (Ty, bool) {
	switch e := e.(type) {
	case *syntax.Ident:
		if t, ok := builtinTypeName(e.Name); ok {
			return t, true
		}
		if t, ok := c.aliasesByName[e.Name]; ok {
			return t, true
		}
	case *syntax.ParenExpr:
		return c.aliasTy(e.X)
	case *syntax.DotExpr:
		if x, ok := e.X.(*syntax.Ident); ok && x.Name == "typing" {
			switch e.Name.Name {
			case "Any":
				return Any(), true
			case "Never":
				return Never(), true
			case "Callable":
				return AnyCallable(), true
			case "Iterable":
				return Iter(Any()), true
			}
		}
	case *syntax.BinaryExpr:
		if e.Op == syntax.PIPE {
			x, okx := c.aliasTy(e.X)
			y, oky := c.aliasTy(e.Y)
			if okx && oky {
				return Union(x, y), true
			}
		}
	case *syntax.IndexExpr:
		// Reuse the lenient interpreter, but require that no
		// approximations were generated.
		before := len(c.o.approx)
		ty := c.o.tyFromIndexExpr(e, c.lookupAnnot)
		if len(c.o.approx) == before {
			return ty, true
		}
		c.o.approx = c.o.approx[:before]
	}
	return Never(), false
}

// lookupAnnot resolves a non-builtin name in annotation position.
func (c *checker) lookupAnnot(name string) (Ty, bool) {
	if t, ok := c.aliasesByName[name]; ok {
		return t, true
	}
	return Never(), false
}

// annotTy interprets a type-annotation expression.
func (c *checker) annotTy(e syntax.Expr) Ty {
	return c.o.tyFromTypeExpr(e, c.lookupAnnot)
}

func (c *checker) collectStmts(stmts []syntax.Stmt) {
	for _, stmt := range stmts {
		c.collectStmt(stmt)
	}
}

func (c *checker) collectStmt(stmt syntax.Stmt) {
	switch stmt := stmt.(type) {
	case *syntax.AssignStmt:
		c.collectExpr(stmt.RHS)
		if stmt.Op == syntax.EQ {
			if stmt.Type != nil {
				want := c.annotTy(stmt.Type)
				c.checkTypes = append(c.checkTypes, checkTypeItem{exprPos(stmt.RHS), stmt.RHS, want})
				if id, ok := unparen(stmt.LHS).(*syntax.Ident); ok {
					c.declare(id, want)
					return
				}
			}
			c.bindTarget(stmt.LHS, exprBind{stmt.RHS})
		} else {
			// augmented assignment
			op := stmt.Op - syntax.PLUS_EQ + syntax.PLUS
			switch lhs := unparen(stmt.LHS).(type) {
			case *syntax.Ident:
				c.bind(lhs, modifyBind{lhs, op, stmt.RHS, stmt.OpPos})
			case *syntax.IndexExpr:
				c.collectExpr(lhs.X)
				c.collectExpr(lhs.Y)
				c.checks = append(c.checks, lhs)
				c.o.approximation("Underapproximation", "augmented assignment to index expression")
			case *syntax.DotExpr:
				c.collectExpr(lhs.X)
				c.o.approximation("Underapproximation", "augmented assignment to attribute")
			}
		}

	case *syntax.DefStmt:
		c.collectDef(stmt)

	case *syntax.ForStmt:
		c.collectExpr(stmt.X)
		c.bindTarget(stmt.Vars, iterBind{exprBind{stmt.X}, exprPos(stmt.X)})
		c.collectStmts(stmt.Body)

	case *syntax.WhileStmt:
		c.collectExpr(stmt.Cond)
		c.checks = append(c.checks, stmt.Cond)
		c.collectStmts(stmt.Body)

	case *syntax.IfStmt:
		c.collectExpr(stmt.Cond)
		c.checks = append(c.checks, stmt.Cond)
		c.collectStmts(stmt.True)
		c.collectStmts(stmt.False)

	case *syntax.ReturnStmt:
		if stmt.Result != nil {
			c.collectExpr(stmt.Result)
		}
		if n := len(c.retStack); n > 0 {
			if want := c.retStack[n-1]; want != nil {
				pos := stmt.Return
				if stmt.Result != nil {
					pos = exprPos(stmt.Result)
				}
				c.checkTypes = append(c.checkTypes, checkTypeItem{pos, stmt.Result, *want})
				return
			}
		}
		if stmt.Result != nil {
			c.checks = append(c.checks, stmt.Result)
		}

	case *syntax.ExprStmt:
		c.collectExpr(stmt.X)
		c.checks = append(c.checks, stmt.X)
		c.collectMutation(stmt.X)

	case *syntax.LoadStmt:
		iface := c.loads[stmt.ModuleName()]
		for i, to := range stmt.To {
			ty, ok := iface.Get(stmt.From[i].Name)
			if !ok {
				c.o.approximation("Unknown load", "%s from %s", stmt.From[i].Name, stmt.ModuleName())
				ty = Any()
			}
			c.declare(to, ty)
		}

	case *syntax.DeferStmt:
		c.collectExpr(stmt.Call)
		c.checks = append(c.checks, stmt.Call)

	case *syntax.ErrDeferStmt:
		c.collectExpr(stmt.Call)
		c.checks = append(c.checks, stmt.Call)

	case *syntax.RecoverStmt:
		c.collectExpr(stmt.Result)
		if n := len(c.catchStack); n > 0 {
			catch := c.catchStack[n-1]
			c.catchTypes[catch] = append(c.catchTypes[catch], stmt.Result)
		}

	case *syntax.BranchStmt:
		// nothing
	}
}

// collectMutation implements the list-mutation special case: a
// statement-level x.append(e) / x.insert(i, e) / x.extend(e) call
// contributes to the type of x (rust's ListAppend/ListExtend).
func (c *checker) collectMutation(e syntax.Expr) {
	call, ok := e.(*syntax.CallExpr)
	if !ok {
		return
	}
	dot, ok := call.Fn.(*syntax.DotExpr)
	if !ok {
		return
	}
	recv, ok := dot.X.(*syntax.Ident)
	if !ok {
		return
	}
	// Positional arguments only.
	for _, arg := range call.Args {
		switch arg := arg.(type) {
		case *syntax.BinaryExpr:
			if arg.Op == syntax.EQ {
				return
			}
		case *syntax.UnaryExpr:
			if arg.Op == syntax.STAR || arg.Op == syntax.STARSTAR {
				return
			}
		}
	}
	switch dot.Name.Name {
	case "append":
		if len(call.Args) == 1 {
			c.bind(recv, appendBind{call.Args[0]})
		}
	case "insert":
		if len(call.Args) == 2 {
			c.bind(recv, appendBind{call.Args[1]})
		}
	case "extend":
		if len(call.Args) == 1 {
			c.bind(recv, extendBind{call.Args[0], exprPos(call.Args[0])})
		}
	}
}

func unparen(e syntax.Expr) syntax.Expr {
	for {
		paren, ok := e.(*syntax.ParenExpr)
		if !ok {
			return e
		}
		e = paren.X
	}
}

// bindTarget attributes the binding expression src to the target(s)
// of an assignment.
func (c *checker) bindTarget(lhs syntax.Expr, src bindExpr) {
	switch lhs := unparen(lhs).(type) {
	case *syntax.Ident:
		c.bind(lhs, src)
	case *syntax.TupleExpr:
		for i, elem := range lhs.List {
			c.bindTarget(elem, indexBind{i, src})
		}
	case *syntax.ListExpr:
		for i, elem := range lhs.List {
			c.bindTarget(elem, indexBind{i, src})
		}
	case *syntax.IndexExpr:
		c.collectExpr(lhs.Y)
		if id, ok := unparen(lhs.X).(*syntax.Ident); ok {
			if eb, ok := src.(exprBind); ok {
				c.bind(id, setIndexBind{id, lhs.Y, eb.e})
				return
			}
		}
		c.collectExpr(lhs.X)
		c.o.approximation("Underapproximation", "assignment to index expression")
		c.checkBindExpr(src)
	case *syntax.DotExpr:
		c.collectExpr(lhs.X)
		c.o.approximation("Underapproximation", "assignment to attribute")
		c.checkBindExpr(src)
	}
}

// checkBindExpr ensures an otherwise-unattached binding expression is
// still type checked for effect.
func (c *checker) checkBindExpr(src bindExpr) {
	if eb, ok := src.(exprBind); ok {
		c.checks = append(c.checks, eb.e)
	}
}

// collectDef processes a def statement: it declares the function's
// own binding with its signature, declares parameter types, and
// collects the body with the declared return type in scope.
func (c *checker) collectDef(stmt *syntax.DefStmt) {
	fn := stmt.Function.(*resolve.Function)

	var params []Param
	addParam := func(id *syntax.Ident, mode ParamMode, required bool, ty Ty) {
		name := ""
		if id != nil {
			name = id.Name
		}
		params = append(params, Param{Name: name, Mode: mode, Required: required, Ty: ty})
	}

	seenStar := false
	mode := func() ParamMode {
		if seenStar {
			return NameOnly
		}
		return PosOrName
	}
	for _, param := range fn.Params {
		var id *syntax.Ident
		var dflt syntax.Expr
		ty := Any()
		var unary *syntax.UnaryExpr

		switch param := param.(type) {
		case *syntax.Ident:
			id = param
		case *syntax.BinaryExpr:
			id = param.X.(*syntax.Ident)
			dflt = param.Y
		case *syntax.UnaryExpr:
			unary = param
		case *syntax.TypedParam:
			ty = c.annotTy(param.Type)
			switch x := param.X.(type) {
			case *syntax.Ident:
				id = x
				dflt = param.Default
			case *syntax.UnaryExpr:
				unary = x
			}
		}

		if unary != nil {
			switch unary.Op {
			case syntax.SLASH:
				// Parameters so far are positional-only.
				for i := range params {
					if params[i].Mode == PosOrName {
						params[i].Mode = PosOnly
					}
				}
			case syntax.STAR:
				seenStar = true
				if uid, _ := unary.X.(*syntax.Ident); uid != nil {
					addParam(uid, ArgsMode, false, ty)
					c.declare(uid, TupleOf(ty))
				}
			default: // **kwargs
				seenStar = true
				if uid, _ := unary.X.(*syntax.Ident); uid != nil {
					addParam(uid, KwargsMode, false, ty)
					c.declare(uid, Dict(Prim("string"), ty))
				}
			}
			continue
		}
		if id != nil {
			addParam(id, mode(), dflt == nil, ty)
			c.declare(id, ty)
			if dflt != nil {
				c.collectExpr(dflt)
				c.checkTypes = append(c.checkTypes, checkTypeItem{exprPos(dflt), dflt, ty})
			}
		}
	}

	result := Any()
	var declaredRet *Ty
	if fn.ReturnType != nil {
		result = c.annotTy(fn.ReturnType)
		declaredRet = &result
	}
	c.declare(stmt.Name, Function(stmt.Name.Name, &ParamSpec{Params: params}, result))

	c.retStack = append(c.retStack, declaredRet)
	c.collectStmts(stmt.Body)
	c.retStack = c.retStack[:len(c.retStack)-1]
}

// collectExpr walks an expression to collect bindings created within
// it: comprehension loop variables, catch-block error variables and
// bodies. Lambdas are not descended into (rust parity).
func (c *checker) collectExpr(e syntax.Expr) {
	if e == nil {
		return
	}
	switch e := e.(type) {
	case *syntax.LambdaExpr:
		c.o.approximation("Lambda", "lambdas are not type checked")

	case *syntax.CatchExpr:
		c.collectExpr(e.X)
		if e.FallbackExpr != nil {
			c.collectExpr(e.FallbackExpr)
		} else {
			c.declare(e.ErrorVar, Prim("error"))
			c.catchStack = append(c.catchStack, e)
			c.collectStmts(e.FallbackBlock)
			c.catchStack = c.catchStack[:len(c.catchStack)-1]
		}

	case *syntax.Comprehension:
		for _, clause := range e.Clauses {
			switch clause := clause.(type) {
			case *syntax.ForClause:
				c.collectExpr(clause.X)
				c.bindTarget(clause.Vars, iterBind{exprBind{clause.X}, exprPos(clause.X)})
			case *syntax.IfClause:
				c.collectExpr(clause.Cond)
			}
		}
		c.collectExpr(e.Body)

	case *syntax.TryExpr:
		c.collectExpr(e.X)

	case *syntax.ParenExpr:
		c.collectExpr(e.X)

	case *syntax.UnaryExpr:
		c.collectExpr(e.X)

	case *syntax.BinaryExpr:
		c.collectExpr(e.X)
		c.collectExpr(e.Y)

	case *syntax.DotExpr:
		c.collectExpr(e.X)

	case *syntax.IndexExpr:
		c.collectExpr(e.X)
		c.collectExpr(e.Y)

	case *syntax.SliceExpr:
		c.collectExpr(e.X)
		c.collectExpr(e.Lo)
		c.collectExpr(e.Hi)
		c.collectExpr(e.Step)

	case *syntax.CondExpr:
		c.collectExpr(e.Cond)
		c.collectExpr(e.True)
		c.collectExpr(e.False)

	case *syntax.CallExpr:
		c.collectExpr(e.Fn)
		for _, arg := range e.Args {
			c.collectExpr(arg)
		}

	case *syntax.TupleExpr:
		for _, x := range e.List {
			c.collectExpr(x)
		}

	case *syntax.ListExpr:
		for _, x := range e.List {
			c.collectExpr(x)
		}

	case *syntax.DictExpr:
		for _, entry := range e.List {
			entry := entry.(*syntax.DictEntry)
			c.collectExpr(entry.Key)
			c.collectExpr(entry.Value)
		}

	case *syntax.Ident, *syntax.Literal, *syntax.EllipsisExpr:
		// leaves
	}
}

// Solving.

const maxIterations = 100

func (c *checker) solve() {
	for _, key := range c.order {
		if d, ok := c.declared[key]; ok {
			c.types[key] = d
		} else {
			c.types[key] = Never()
		}
	}
	c.o.suppress = true
	for i := 0; ; i++ {
		if i == maxIterations {
			c.o.suppress = false
			c.o.approximation("Fixed point", "solver did not converge in %d iterations", maxIterations)
			break
		}
		changed := false
		for _, key := range c.order {
			t := c.types[key]
			for _, be := range c.exprs[key] {
				t = Union(t, be.ty(c))
			}
			if !t.Equal(c.types[key]) {
				c.types[key] = t
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	c.o.suppress = false
}

// check runs the final, error-reporting pass: every binding
// expression and deferred check is evaluated once under the solved
// types.
func (c *checker) check() {
	for _, key := range c.order {
		for _, be := range c.exprs[key] {
			be.ty(c)
		}
	}
	for _, e := range c.checks {
		c.exprType(e)
	}
	for _, item := range c.checkTypes {
		got := None() // bare `return`
		if item.expr != nil {
			got = c.exprType(item.expr)
		}
		c.o.validateType(item.pos, got, item.want)
	}
}

func (c *checker) typeOfBinding(b *resolve.Binding) Ty {
	if t, ok := c.types[bindKeyOf(b)]; ok {
		return t
	}
	return Any()
}

func (c *checker) typeMap() *TypeMap {
	tm := &TypeMap{m: make(map[any]bindingInfo, len(c.infos))}
	for key, info := range c.infos {
		info.ty = c.types[key]
		tm.m[key] = info
	}
	return tm
}

// Expression inference.

func (c *checker) exprType(e syntax.Expr) Ty {
	switch e := e.(type) {
	case *syntax.Ident:
		return c.identType(e)

	case *syntax.Literal:
		switch e.Token {
		case syntax.INT:
			return Prim("int")
		case syntax.FLOAT:
			return Prim("float")
		case syntax.STRING:
			return Prim("string")
		case syntax.BYTES:
			return Prim("bytes")
		}
		return Any()

	case *syntax.ParenExpr:
		return c.exprType(e.X)

	case *syntax.ListExpr:
		elems := make([]Ty, len(e.List))
		for i, x := range e.List {
			elems[i] = c.exprType(x)
		}
		return List(Union(elems...))

	case *syntax.TupleExpr:
		elems := make([]Ty, len(e.List))
		for i, x := range e.List {
			elems[i] = c.exprType(x)
		}
		return Tuple(elems...)

	case *syntax.DictExpr:
		var keys, vals []Ty
		for _, entry := range e.List {
			entry := entry.(*syntax.DictEntry)
			keys = append(keys, c.exprType(entry.Key))
			vals = append(vals, c.exprType(entry.Value))
		}
		return Dict(Union(keys...), Union(vals...))

	case *syntax.CondExpr:
		c.exprType(e.Cond)
		return Union(c.exprType(e.True), c.exprType(e.False))

	case *syntax.UnaryExpr:
		return c.o.unop(e.OpPos, e.Op, c.exprType(e.X))

	case *syntax.BinaryExpr:
		return c.binaryType(e)

	case *syntax.DotExpr:
		return c.o.attr(e.Dot, c.exprType(e.X), e.Name.Name)

	case *syntax.IndexExpr:
		return c.o.index(e.Lbrack, c.exprType(e.X), c.exprType(e.Y))

	case *syntax.SliceExpr:
		intOrNone := Union(Prim("int"), None())
		for _, bound := range []syntax.Expr{e.Lo, e.Hi, e.Step} {
			if bound != nil {
				c.o.validateType(exprPos(bound), c.exprType(bound), intOrNone)
			}
		}
		return c.o.slice(e.Lbrack, c.exprType(e.X))

	case *syntax.CallExpr:
		return c.callType(e)

	case *syntax.LambdaExpr:
		return AnyCallable()

	case *syntax.Comprehension:
		for _, clause := range e.Clauses {
			if ifc, ok := clause.(*syntax.IfClause); ok {
				c.exprType(ifc.Cond)
			}
		}
		if e.Curly {
			entry := e.Body.(*syntax.DictEntry)
			return Dict(c.exprType(entry.Key), c.exprType(entry.Value))
		}
		return List(c.exprType(e.Body))

	case *syntax.TryExpr:
		return c.exprType(e.X)

	case *syntax.CatchExpr:
		xt := c.exprType(e.X)
		if e.FallbackExpr != nil {
			return Union(xt, c.exprType(e.FallbackExpr))
		}
		recovers := c.catchTypes[e]
		tys := []Ty{xt}
		for _, r := range recovers {
			tys = append(tys, c.exprType(r))
		}
		return Union(tys...)

	case *syntax.EllipsisExpr:
		return Any()
	}
	return Any()
}

func (c *checker) identType(id *syntax.Ident) Ty {
	b, ok := id.Binding.(*resolve.Binding)
	if !ok || b == nil {
		return Any()
	}
	switch b.Scope {
	case resolve.Universal:
		if t, ok := universeTypes[id.Name]; ok {
			return t
		}
		return Any()
	case resolve.Predeclared:
		if t, ok := c.env[id.Name]; ok {
			return t
		}
		return Any()
	default:
		if t, ok := c.types[bindKeyOf(b)]; ok {
			return t
		}
		return Any()
	}
}

func (c *checker) binaryType(e *syntax.BinaryExpr) Ty {
	switch e.Op {
	case syntax.AND, syntax.OR:
		return Union(c.exprType(e.X), c.exprType(e.Y))
	case syntax.EQL, syntax.NEQ:
		c.exprType(e.X)
		c.exprType(e.Y)
		return Prim("bool")
	case syntax.IN:
		return c.o.contains(e.OpPos, c.exprType(e.X), c.exprType(e.Y))
	case syntax.NOT_IN:
		c.o.contains(e.OpPos, c.exprType(e.X), c.exprType(e.Y))
		return Prim("bool")
	default:
		return c.o.binop(e.OpPos, e.Op, c.exprType(e.X), c.exprType(e.Y))
	}
}

func (c *checker) callType(e *syntax.CallExpr) Ty {
	var args callArgs
	for _, arg := range e.Args {
		switch arg := arg.(type) {
		case *syntax.BinaryExpr:
			if arg.Op == syntax.EQ {
				if id, ok := arg.X.(*syntax.Ident); ok {
					args.named = append(args.named, namedArg{id.Name, c.exprType(arg.Y), exprPos(arg)})
					continue
				}
			}
			args.pos = append(args.pos, tyPos{c.exprType(arg), exprPos(arg)})
		case *syntax.UnaryExpr:
			switch arg.Op {
			case syntax.STAR:
				t := c.exprType(arg.X)
				c.o.iterItem(exprPos(arg), t) // *args must be iterable
				args.starArgs = &tyPos{t, exprPos(arg)}
				continue
			case syntax.STARSTAR:
				t := c.exprType(arg.X)
				c.o.validateType(exprPos(arg), t, Dict(Prim("string"), Any()))
				args.kwargsArgs = &tyPos{t, exprPos(arg)}
				continue
			}
			args.pos = append(args.pos, tyPos{c.exprType(arg), exprPos(arg)})
		default:
			args.pos = append(args.pos, tyPos{c.exprType(arg), exprPos(arg)})
		}
	}
	return c.o.call(exprPos(e), c.exprType(e.Fn), args)
}
