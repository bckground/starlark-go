// Copyright 2026 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typecheck

// The oracle answers questions about the static types of operations:
// attribute access, indexing, unary and binary operators, calls, and
// iteration. It is the Go analogue of starlark-rust's TypingOracleCtx
// (starlark/src/typing/oracle/ctx.rs).
//
// Leniency principle: when the oracle does not know, it answers Any
// and never reports an error. Errors are reserved for operations that
// are certainly invalid for every alternative of a union.

import (
	"fmt"

	"go.starlark.net/syntax"
)

// intersects reports whether a value could simultaneously inhabit
// both types. Compatibility checking uses intersection, not subtyping
// (like starlark-rust): a check fails only when the types certainly
// do not overlap.
func intersects(a, b Ty) bool {
	if a.IsAny() || b.IsAny() {
		return true
	}
	// Never intersects everything: it has no values, so no evidence
	// of a conflict can exist (matches rust's "Never" leniency).
	if a.IsNever() || b.IsNever() {
		return true
	}
	for _, x := range a.alts {
		for _, y := range b.alts {
			if basicIntersect(x, y) {
				return true
			}
		}
	}
	return false
}

func basicIntersect(x, y Basic) bool {
	if _, ok := x.(anyBasic); ok {
		return true
	}
	if _, ok := y.(anyBasic); ok {
		return true
	}
	// An iterable intersects anything that can be iterated.
	if it, ok := x.(iterBasic); ok {
		return iterableIntersect(it, y)
	}
	if it, ok := y.(iterBasic); ok {
		return iterableIntersect(it, x)
	}
	switch x := x.(type) {
	case primBasic:
		y, ok := y.(primBasic)
		if !ok {
			return false
		}
		if x.name == y.name {
			return true
		}
		// Numeric coercion: a float annotation accepts ints at
		// runtime (starlark-rust parity), so int and float intersect.
		isNum := func(name string) bool { return name == "int" || name == "float" }
		return isNum(x.name) && isNum(y.name)
	case listBasic:
		y, ok := y.(listBasic)
		return ok && intersects(x.elem, y.elem)
	case setBasic:
		y, ok := y.(setBasic)
		return ok && intersects(x.elem, y.elem)
	case dictBasic:
		y, ok := y.(dictBasic)
		return ok && intersects(x.key, y.key) && intersects(x.val, y.val)
	case tupleBasic:
		y, ok := y.(tupleBasic)
		return ok && tupleIntersect(x, y)
	case callableBasic:
		y, ok := y.(callableBasic)
		return ok && callablesIntersect(x, y)
	case typeBasic:
		_, ok := y.(typeBasic)
		return ok
	case moduleBasic:
		_, ok := y.(moduleBasic)
		return ok
	case customBasic:
		// Nominal identity: only the same minted type intersects
		// (like the runtime's pointer-identity TypeMatcher).
		y, ok := y.(customBasic)
		return ok && x.serial == y.serial
	}
	return false
}

// callablesIntersect reports whether two callable types intersect:
// their parameters must be compatible and their results must
// intersect (rust's callables_intersect).
func callablesIntersect(x, y callableBasic) bool {
	return paramsIntersect(x.params, y.params) && intersects(x.result, y.result)
}

// paramsIntersect reports whether two parameter specifications are
// compatible, a port of rust's params_intersect. The precise check
// applies only when at least one spec is in "simple" form — all
// parameters required, positional-only or named-only, the shape of a
// typing.Callable[[...], R] annotation. Two general specs always
// intersect (leniency).
func paramsIntersect(a, b *ParamSpec) bool {
	if a.IsAny() || b.IsAny() {
		return true
	}
	apos, anamed, aSimple := allRequiredPosOnlyNamedOnly(a)
	bpos, bnamed, bSimple := allRequiredPosOnlyNamedOnly(b)
	switch {
	case aSimple && bSimple:
		if len(apos) != len(bpos) || len(anamed) != len(bnamed) {
			return false
		}
		for i := range apos {
			if !intersects(apos[i], bpos[i]) {
				return false
			}
		}
		byName := make(map[string]Ty, len(bnamed))
		for _, p := range bnamed {
			byName[p.Name] = p.Ty
		}
		for _, p := range anamed {
			q, ok := byName[p.Name]
			if !ok || !intersects(p.Ty, q) {
				return false
			}
		}
		return true
	case aSimple:
		return simpleParamsCallIntersect(apos, anamed, b)
	case bSimple:
		return simpleParamsCallIntersect(bpos, bnamed, a)
	default:
		return true
	}
}

// allRequiredPosOnlyNamedOnly decomposes a spec into its positional-only
// and named-only parameter types if every parameter is required and of
// one of those modes (rust's all_required_pos_only_named_only).
func allRequiredPosOnlyNamedOnly(ps *ParamSpec) (pos []Ty, named []Param, ok bool) {
	for _, p := range ps.Params {
		switch {
		case p.Mode == PosOnly && p.Required:
			pos = append(pos, p.Ty)
		case p.Mode == NameOnly && p.Required:
			named = append(named, p)
		default:
			return nil, nil, false
		}
	}
	return pos, named, true
}

// simpleParamsCallIntersect reports whether a call with the given
// required positional and named argument types would be accepted by
// spec (rust's params_all_pos_only_named_only_intersect).
func simpleParamsCallIntersect(pos []Ty, named []Param, spec *ParamSpec) bool {
	var args callArgs
	for _, t := range pos {
		args.pos = append(args.pos, tyPos{ty: t})
	}
	for _, p := range named {
		args.named = append(args.named, namedArg{name: p.Name, ty: p.Ty})
	}
	var o oracle
	_, callErr := o.validateCall(syntax.Position{}, callableBasic{params: spec, result: Any()}, args)
	return callErr == nil
}

func tupleIntersect(x, y tupleBasic) bool {
	if x.of != nil && y.of != nil {
		return intersects(*x.of, *y.of)
	}
	if x.of != nil {
		x, y = y, x // x fixed, y open
	}
	if y.of != nil {
		for _, e := range x.elems {
			if !intersects(e, *y.of) {
				return false
			}
		}
		return true
	}
	if len(x.elems) != len(y.elems) {
		return false
	}
	for i := range x.elems {
		if !intersects(x.elems[i], y.elems[i]) {
			return false
		}
	}
	return true
}

func iterableIntersect(it iterBasic, y Basic) bool {
	item, ok := iterItemBasic(y)
	if !ok {
		return false
	}
	return intersects(it.item, item)
}

// iterItemBasic returns the element type yielded by iterating a value
// of basic type b, if b is iterable.
func iterItemBasic(b Basic) (Ty, bool) {
	switch b := b.(type) {
	case anyBasic:
		return Any(), true
	case listBasic:
		return b.elem, true
	case setBasic:
		return b.elem, true
	case dictBasic:
		return b.key, true
	case tupleBasic:
		if b.of != nil {
			return *b.of, true
		}
		return Union(b.elems...), true
	case iterBasic:
		return b.item, true
	case primBasic:
		switch b.name {
		case "range":
			return Prim("int"), true
		case "bytes":
			return Prim("int"), true
		}
	case customBasic:
		if it, ok := b.c.(CustomIterable); ok {
			return it.IterItem(), true
		}
	}
	return Never(), false
}

// oracle carries the context of one Check invocation.
type oracle struct {
	errors   []Error
	approx   []Approximation
	typeMap  *TypeMap
	suppress bool // do not record errors (during fixpoint iteration)
}

func (o *oracle) errorf(pos syntax.Position, format string, args ...any) {
	if o.suppress {
		return
	}
	o.errors = append(o.errors, Error{Pos: pos, Msg: fmt.Sprintf(format, args...)})
}

func (o *oracle) approximation(category, format string, args ...any) {
	if o.suppress {
		return
	}
	o.approx = append(o.approx, Approximation{Category: category, Message: fmt.Sprintf(format, args...)})
}

// validateType checks that got is compatible with (intersects) want,
// reporting an error at pos otherwise.
func (o *oracle) validateType(pos syntax.Position, got, want Ty) {
	if !intersects(got, want) {
		o.errorf(pos, "Expected type `%s` but got `%s`", want, got)
	}
}

// iterItem returns the type of elements yielded by iterating ty.
func (o *oracle) iterItem(pos syntax.Position, ty Ty) Ty {
	if ty.IsAny() || ty.IsNever() {
		return ty
	}
	var items []Ty
	for _, alt := range ty.alts {
		if item, ok := iterItemBasic(alt); ok {
			items = append(items, item)
		}
	}
	if items == nil {
		o.errorf(pos, "Type `%s` is not iterable", ty)
		return Never()
	}
	return Union(items...)
}

// attr returns the type of expression x.name where x has type ty.
func (o *oracle) attr(pos syntax.Position, ty Ty, name string) Ty {
	if ty.IsAny() || ty.IsNever() {
		return ty
	}
	var results []Ty
	for _, alt := range ty.alts {
		if t, ok := basicAttr(alt, name); ok {
			results = append(results, t)
		}
	}
	if results == nil {
		o.errorf(pos, "Object of type `%s` has no attribute `%s`", ty, name)
		return Never()
	}
	return Union(results...)
}

func basicAttr(b Basic, name string) (Ty, bool) {
	switch b := b.(type) {
	case anyBasic:
		return Any(), true
	case listBasic:
		return listMethod(b.elem, name)
	case dictBasic:
		return dictMethod(b.key, b.val, name)
	case setBasic:
		return setMethod(b.elem, name)
	case primBasic:
		switch b.name {
		case "string":
			return stringMethod(name)
		case "bytes":
			if name == "elems" {
				return Function("elems", PositionalOnly(), Iter(Prim("int"))), true
			}
		case "error":
			// fork-specific: error values; attributes unknown
			return Any(), true
		case "error_tag":
			return Any(), true
		}
	case moduleBasic:
		if t, ok := b.attrs[name]; ok {
			return t, true
		}
		// Unknown module member: lenient.
		return Any(), true
	case customBasic:
		return b.c.Attr(name)
	case callableBasic, tupleBasic, iterBasic, typeBasic:
		// no attributes
	}
	return Never(), false
}

// unop returns the type of a unary operation.
func (o *oracle) unop(pos syntax.Position, op syntax.Token, ty Ty) Ty {
	if op == syntax.NOT {
		if ty.IsNever() {
			return Never()
		}
		return Prim("bool")
	}
	if ty.IsAny() || ty.IsNever() {
		return ty
	}
	var results []Ty
	for _, alt := range ty.alts {
		if prim, ok := alt.(primBasic); ok {
			switch prim.name {
			case "int":
				if op == syntax.MINUS || op == syntax.PLUS || op == syntax.TILDE {
					results = append(results, Prim("int"))
				}
			case "float":
				if op == syntax.MINUS || op == syntax.PLUS {
					results = append(results, Prim("float"))
				}
			}
		}
	}
	if results == nil {
		o.errorf(pos, "Unary operator `%s` is not available on the type `%s`", op, ty)
		return Never()
	}
	return Union(results...)
}

// binop returns the type of a binary operation l op r.
// A union operand succeeds if any pair of alternatives succeeds.
func (o *oracle) binop(pos syntax.Position, op syntax.Token, l, r Ty) Ty {
	if l.IsAny() || r.IsAny() {
		return Any()
	}
	if l.IsNever() || r.IsNever() {
		return Never()
	}
	var results []Ty
	for _, x := range l.alts {
		for _, y := range r.alts {
			if t, ok := basicBinop(op, x, y); ok {
				results = append(results, t)
			}
		}
	}
	if results == nil {
		o.errorf(pos, "Binary operator `%s` is not available on the types `%s` and `%s`", op, l, r)
		return Never()
	}
	return Union(results...)
}

func basicBinop(op syntax.Token, x, y Basic) (Ty, bool) {
	// Custom types may define binary operators the checker cannot
	// see (e.g. error tag sets merge with |): lenient.
	if _, ok := x.(customBasic); ok {
		return Any(), true
	}
	if _, ok := y.(customBasic); ok {
		return Any(), true
	}

	xp, xIsPrim := x.(primBasic)
	yp, yIsPrim := y.(primBasic)

	// numeric arithmetic
	isNum := func(p primBasic) bool { return p.name == "int" || p.name == "float" }
	if xIsPrim && yIsPrim && isNum(xp) && isNum(yp) {
		switch op {
		case syntax.PLUS, syntax.MINUS, syntax.STAR, syntax.PERCENT, syntax.SLASHSLASH:
			if xp.name == "float" || yp.name == "float" {
				return Prim("float"), true
			}
			return Prim("int"), true
		case syntax.SLASH:
			return Prim("float"), true
		case syntax.LT, syntax.GT, syntax.LE, syntax.GE:
			return Prim("bool"), true
		}
		// bitwise and shifts: ints only
		if xp.name == "int" && yp.name == "int" {
			switch op {
			case syntax.AMP, syntax.PIPE, syntax.CIRCUMFLEX, syntax.LTLT, syntax.GTGT:
				return Prim("int"), true
			}
		}
		return Never(), false
	}

	switch op {
	case syntax.PLUS:
		switch x := x.(type) {
		case primBasic:
			if yIsPrim && xp.name == yp.name && (xp.name == "string" || xp.name == "bytes") {
				return Prim(xp.name), true
			}
		case listBasic:
			if y, ok := y.(listBasic); ok {
				return List(Union(x.elem, y.elem)), true
			}
		case tupleBasic:
			if _, ok := y.(tupleBasic); ok {
				return AnyTuple(), true
			}
		}

	case syntax.STAR:
		// sequence repetition
		if yIsPrim && yp.name == "int" {
			switch x := x.(type) {
			case primBasic:
				if x.name == "string" || x.name == "bytes" {
					return Prim(x.name), true
				}
			case listBasic:
				return Ty{alts: []Basic{x}}, true
			case tupleBasic:
				return Ty{alts: []Basic{x}}, true
			}
		}
		if xIsPrim && xp.name == "int" {
			switch y := y.(type) {
			case primBasic:
				if y.name == "string" || y.name == "bytes" {
					return Prim(y.name), true
				}
			case listBasic:
				return Ty{alts: []Basic{y}}, true
			case tupleBasic:
				return Ty{alts: []Basic{y}}, true
			}
		}

	case syntax.PERCENT:
		// string formatting: str % anything -> str
		if xIsPrim && xp.name == "string" {
			return Prim("string"), true
		}

	case syntax.PIPE:
		switch x := x.(type) {
		case dictBasic:
			if y, ok := y.(dictBasic); ok {
				return Dict(Union(x.key, y.key), Union(x.val, y.val)), true
			}
		case setBasic:
			if y, ok := y.(setBasic); ok {
				return Set(Union(x.elem, y.elem)), true
			}
		}
		// type union expression: int | None
		if isTypeLikeBasic(x) && isTypeLikeBasic(y) {
			return TypeType(), true
		}

	case syntax.AMP, syntax.CIRCUMFLEX:
		if x, ok := x.(setBasic); ok {
			if y, ok := y.(setBasic); ok {
				return Set(Union(x.elem, y.elem)), true
			}
		}

	case syntax.MINUS:
		if x, ok := x.(setBasic); ok {
			if _, ok := y.(setBasic); ok {
				return Ty{alts: []Basic{x}}, true
			}
		}

	case syntax.LT, syntax.GT, syntax.LE, syntax.GE:
		// ordered comparison of like types
		if x.basicSortKey() == y.basicSortKey() {
			return Prim("bool"), true
		}
		switch x.(type) {
		case listBasic:
			if _, ok := y.(listBasic); ok {
				return Prim("bool"), true
			}
		case tupleBasic:
			if _, ok := y.(tupleBasic); ok {
				return Prim("bool"), true
			}
		case primBasic:
			if yIsPrim && xp.name == yp.name {
				return Prim("bool"), true
			}
		}
	}
	return Never(), false
}

// contains returns the type of `needle in haystack`.
func (o *oracle) contains(pos syntax.Position, needle, haystack Ty) Ty {
	if haystack.IsAny() || haystack.IsNever() || needle.IsAny() || needle.IsNever() {
		return Prim("bool")
	}
	ok := false
	for _, alt := range haystack.alts {
		switch alt := alt.(type) {
		case dictBasic:
			if intersects(needle, alt.key) {
				ok = true
			}
		case primBasic:
			if alt.name == "string" && intersects(needle, Prim("string")) {
				ok = true
			}
			if alt.name == "bytes" && intersects(needle, Union(Prim("bytes"), Prim("int"))) {
				ok = true
			}
			if alt.name == "range" && intersects(needle, Prim("int")) {
				ok = true
			}
		default:
			if item, isIter := iterItemBasic(alt); isIter {
				if intersects(needle, item) {
					ok = true
				}
			}
		}
	}
	if !ok {
		o.errorf(pos, "Binary operator `in` is not available on the types `%s` and `%s`", needle, haystack)
		return Never()
	}
	return Prim("bool")
}

// index returns the type of x[i].
func (o *oracle) index(pos syntax.Position, ty, index Ty) Ty {
	if ty.IsAny() || ty.IsNever() {
		return ty
	}
	var results []Ty
	for _, alt := range ty.alts {
		switch alt := alt.(type) {
		case anyBasic:
			results = append(results, Any())
		case listBasic:
			if intersects(index, Prim("int")) {
				results = append(results, alt.elem)
			}
		case dictBasic:
			if intersects(index, alt.key) {
				results = append(results, alt.val)
			}
		case tupleBasic:
			if intersects(index, Prim("int")) {
				if alt.of != nil {
					results = append(results, *alt.of)
				} else {
					results = append(results, Union(alt.elems...))
				}
			}
		case primBasic:
			switch alt.name {
			case "string":
				if intersects(index, Prim("int")) {
					results = append(results, Prim("string"))
				}
			case "bytes", "range":
				if intersects(index, Prim("int")) {
					results = append(results, Prim("int"))
				}
			}
		case typeBasic:
			// parameterization of a type value: list[int]
			results = append(results, TypeType())
		case callableBasic:
			// The canonical type constructors are first-class type
			// values at runtime: list[int] parameterizes them.
			if isTypeConstructorName(alt.name) {
				results = append(results, TypeType())
			}
		case customBasic:
			if ix, ok := alt.c.(CustomIndexable); ok {
				if t, ok := ix.IndexResult(index); ok {
					results = append(results, t)
				}
			}
		}
	}
	if results == nil {
		o.errorf(pos, "Type `%s` does not have [] operator on `%s`", ty, index)
		return Never()
	}
	return Union(results...)
}

// isTypeConstructorName reports whether name is a universal builtin
// that doubles as a type value (see starlark.TypeOf).
func isTypeConstructorName(name string) bool {
	switch name {
	case "bool", "bytes", "dict", "float", "int", "list", "range", "set", "str", "tuple", "type":
		return true
	}
	return false
}

// isTypeLikeBasic reports whether a value of this basic type can act
// as a type in a union expression (x | y).
func isTypeLikeBasic(b Basic) bool {
	switch b := b.(type) {
	case typeBasic:
		return true
	case primBasic:
		return b.name == "NoneType"
	case callableBasic:
		return isTypeConstructorName(b.name)
	}
	return false
}

// slice returns the type of x[lo:hi:step].
func (o *oracle) slice(pos syntax.Position, ty Ty) Ty {
	if ty.IsAny() || ty.IsNever() {
		return ty
	}
	var results []Ty
	for _, alt := range ty.alts {
		switch alt.(type) {
		case listBasic, tupleBasic, primBasic:
			if prim, ok := alt.(primBasic); ok {
				switch prim.name {
				case "string", "bytes", "range":
					// sliceable
				default:
					continue
				}
			}
			results = append(results, Ty{alts: []Basic{alt}})
		}
	}
	if results == nil {
		o.errorf(pos, "Type `%s` does not have [::] operator", ty)
		return Never()
	}
	return Union(results...)
}

// callArgs describes the arguments of a call expression.
type callArgs struct {
	pos        []tyPos    // positional arguments
	named      []namedArg // named arguments
	starArgs   *tyPos     // *args, if any
	kwargsArgs *tyPos     // **kwargs, if any
}

type tyPos struct {
	ty  Ty
	pos syntax.Position
}

type namedArg struct {
	name string
	ty   Ty
	pos  syntax.Position
}

// call returns the type of a call to a value of type fnTy.
func (o *oracle) call(pos syntax.Position, fnTy Ty, args callArgs) Ty {
	if fnTy.IsAny() || fnTy.IsNever() {
		return fnTy
	}
	var results []Ty
	var firstErr *Error
	for _, alt := range fnTy.alts {
		switch alt := alt.(type) {
		case anyBasic:
			results = append(results, Any())
		case callableBasic:
			t, callErr := o.validateCall(pos, alt, args)
			if callErr != nil {
				if firstErr == nil {
					firstErr = callErr
				}
				continue
			}
			if alt.specialFn != nil {
				if st, ok := alt.specialFn(o, pos, args); ok {
					t = st
				}
			}
			results = append(results, t)
		case typeBasic:
			// calling a type value constructs an instance: unknown
			results = append(results, Any())
		case primBasic:
			if alt.name == "error_tag" {
				// Error tags are callable, constructing an error:
				// NotFound(message="...").
				results = append(results, Prim("error"))
			}
		case customBasic:
			if cc, ok := alt.c.(CustomCallable); ok {
				params, result := cc.CallSignature()
				t, callErr := o.validateCall(pos, callableBasic{name: alt.c.TyName(), params: params, result: result}, args)
				if callErr != nil {
					if firstErr == nil {
						firstErr = callErr
					}
					continue
				}
				results = append(results, t)
			}
		}
	}
	if results == nil {
		if firstErr != nil {
			o.errorf(firstErr.Pos, "%s", firstErr.Msg)
		} else {
			o.errorf(pos, "Call to a non-callable type `%s`", fnTy)
		}
		return Never()
	}
	return Union(results...)
}

// validateCall checks arguments against a callable signature and
// returns the result type, or an error.
func (o *oracle) validateCall(pos syntax.Position, fn callableBasic, args callArgs) (Ty, *Error) {
	if fn.params.IsAny() {
		return fn.result, nil
	}
	errf := func(p syntax.Position, format string, a ...any) (Ty, *Error) {
		return Never(), &Error{Pos: p, Msg: fmt.Sprintf(format, a...)}
	}

	params := fn.params.Params
	filled := make([]bool, len(params))

	var argsParam, kwargsParam *Param
	for i := range params {
		switch params[i].Mode {
		case ArgsMode:
			argsParam = &params[i]
		case KwargsMode:
			kwargsParam = &params[i]
		}
	}

	// positional arguments
	pi := 0
	for _, arg := range args.pos {
		// find next fillable positional parameter
		for pi < len(params) && (params[pi].Mode == NameOnly || params[pi].Mode == ArgsMode || params[pi].Mode == KwargsMode || filled[pi]) {
			if params[pi].Mode == NameOnly || params[pi].Mode == KwargsMode {
				pi = len(params)
				break
			}
			pi++
		}
		if pi >= len(params) || !(params[pi].Mode == PosOnly || params[pi].Mode == PosOrName) {
			if argsParam != nil {
				if !intersects(arg.ty, argsParam.Ty) {
					return errf(arg.pos, "Expected type `%s` but got `%s`", argsParam.Ty, arg.ty)
				}
				continue
			}
			return errf(arg.pos, "Too many positional arguments")
		}
		if !intersects(arg.ty, params[pi].Ty) {
			return errf(arg.pos, "Expected type `%s` but got `%s`", params[pi].Ty, arg.ty)
		}
		filled[pi] = true
		pi++
	}

	// named arguments
	for _, arg := range args.named {
		found := false
		for i := range params {
			if params[i].Name == arg.name && (params[i].Mode == PosOrName || params[i].Mode == NameOnly) {
				if filled[i] {
					return errf(arg.pos, "Parameter `%s` occurs both explicitly and in **kwargs", arg.name)
				}
				if !intersects(arg.ty, params[i].Ty) {
					return errf(arg.pos, "Expected type `%s` but got `%s`", params[i].Ty, arg.ty)
				}
				filled[i] = true
				found = true
				break
			}
		}
		if !found {
			if kwargsParam != nil {
				if !intersects(arg.ty, kwargsParam.Ty) {
					return errf(arg.pos, "Expected type `%s` but got `%s`", kwargsParam.Ty, arg.ty)
				}
				continue
			}
			return errf(arg.pos, "Unexpected parameter named `%s`", arg.name)
		}
	}

	// required parameters: suppressed when *args/**kwargs are present
	// at the call site, which could fill them (rust's seen_vargs rule).
	if args.starArgs == nil && args.kwargsArgs == nil {
		for i := range params {
			if params[i].Required && !filled[i] &&
				params[i].Mode != ArgsMode && params[i].Mode != KwargsMode {
				name := params[i].Name
				if name == "" {
					name = fmt.Sprintf("%d", i)
				}
				return errf(pos, "Missing required parameter `%s`", name)
			}
		}
	}
	return fn.result, nil
}
