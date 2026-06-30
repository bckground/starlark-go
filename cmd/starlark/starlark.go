// Copyright 2017 The Bazel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The starlark command interprets a Starlark file.
// With no arguments, it starts a read-eval-print loop (REPL).
package main // import "go.starlark.net/cmd/starlark"

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	"go.starlark.net/internal/compile"
	"go.starlark.net/lib/json"
	"go.starlark.net/lib/math"
	"go.starlark.net/lib/time"
	"go.starlark.net/lib/typing"
	"go.starlark.net/repl"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
	"go.starlark.net/typecheck"
	"golang.org/x/term"
)

// flags
var (
	cpuprofile = flag.String("cpuprofile", "", "gather Go CPU profile in this file")
	memprofile = flag.String("memprofile", "", "gather Go memory profile in this file")
	profile    = flag.String("profile", "", "gather Starlark time profile in this file")
	showenv    = flag.Bool("showenv", false, "on success, print final global environment")
	execprog   = flag.String("c", "", "execute program `prog`")
	types      = flag.String("types", "off", `support for type annotations: "off", "parse", or "on"`)
	typecheckF = flag.Bool("typecheck", false, "run the static typechecker before execution (implies -types=on unless set)")
	posonly    = flag.Bool("positionalonly", false, "allow positional-only parameters: def f(x, /)")
)

func init() {
	flag.BoolVar(&compile.Disassemble, "disassemble", compile.Disassemble, "show disassembly during compilation of each function")

	// non-standard dialect flags
	flag.BoolVar(&resolve.AllowRecursion, "recursion", resolve.AllowRecursion, "allow while statements and recursive functions")
	flag.BoolVar(&resolve.AllowGlobalReassign, "globalreassign", resolve.AllowGlobalReassign, "allow reassignment of globals, and if/for/while statements at top level")

	// obsolete flags for features that are now standard
	flag.BoolVar(&resolve.AllowSet, "set", true, "obsolete; no effect")
	flag.BoolVar(&resolve.AllowFloat, "float", true, "obsolete; no effect")
	flag.BoolVar(&resolve.AllowLambda, "lambda", true, "obsolete; no effect")
}

func main() {
	os.Exit(doMain())
}

func doMain() int {
	log.SetPrefix("starlark: ")
	log.SetFlags(0)
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		check(err)
		err = pprof.StartCPUProfile(f)
		check(err)
		defer func() {
			pprof.StopCPUProfile()
			err := f.Close()
			check(err)
		}()
	}
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		check(err)
		defer func() {
			runtime.GC()
			err := pprof.Lookup("heap").WriteTo(f, 0)
			check(err)
			err = f.Close()
			check(err)
		}()
	}

	if *profile != "" {
		f, err := os.Create(*profile)
		check(err)
		err = starlark.StartProfile(f)
		check(err)
		defer func() {
			err := starlark.StopProfile()
			check(err)
		}()
	}

	opts := syntax.LegacyFileOptions()
	switch *types {
	case "off":
		opts.Types = syntax.TypesDisabled
	case "parse":
		opts.Types = syntax.TypesParseOnly
	case "on":
		opts.Types = syntax.TypesEnabled
	default:
		log.Printf(`invalid -types value %q: want "off", "parse", or "on"`, *types)
		return 1
	}
	if *typecheckF && opts.Types == syntax.TypesDisabled {
		opts.Types = syntax.TypesEnabled
	}
	opts.PositionalOnly = *posonly

	thread := &starlark.Thread{Load: repl.MakeLoadOptions(opts)}
	globals := make(starlark.StringDict)

	// Ideally this statement would update the predeclared environment.
	// TODO(adonovan): plumb predeclared env through to the REPL.
	starlark.Universe["json"] = json.Module
	starlark.Universe["time"] = time.Module
	starlark.Universe["math"] = math.Module
	starlark.Universe["typing"] = typing.Module

	switch {
	case flag.NArg() == 1 || *execprog != "":
		var (
			filename string
			src      any
			err      error
		)
		if *execprog != "" {
			// Execute provided program.
			filename = "cmdline"
			src = *execprog
		} else {
			// Execute specified file.
			filename = flag.Arg(0)
		}
		thread.Name = "exec " + filename
		if *typecheckF {
			// The exec path passes no predeclared dict, so nothing is
			// predeclared here either; names like print resolve as
			// universal (the modules below were added to the Universe).
			f, prog, err := starlark.SourceProgramOptions(opts, filename, src, noPredeclared)
			if err != nil {
				repl.PrintError(err)
				return 1
			}
			env := typecheck.UniverseEnv()
			for _, mod := range []string{"json", "time", "math", "typing"} {
				env[mod] = typecheck.Module(mod, nil)
			}
			lc := &loadChecker{opts: opts, env: env, cache: make(map[string]*typecheck.Interface)}
			lc.cache[filename] = nil // loading the main module again is a cycle
			if _, err := lc.checkFile(f); err != nil {
				log.Print(err)
				return 1
			}
			for _, e := range lc.errors {
				fmt.Fprintln(os.Stderr, e.Error())
			}
			if len(lc.errors) > 0 {
				return 1
			}
			globals, err = prog.Init(thread, nil)
			if err != nil {
				repl.PrintError(err)
				return 1
			}
		} else {
			globals, err = starlark.ExecFileOptions(opts, thread, filename, src, nil)
			if err != nil {
				repl.PrintError(err)
				return 1
			}
		}
	case flag.NArg() == 0:
		stdinIsTerminal := term.IsTerminal(int(os.Stdin.Fd()))
		if stdinIsTerminal {
			fmt.Println("Welcome to Starlark (go.starlark.net)")
		}
		thread.Name = "REPL"
		repl.REPLOptions(opts, thread, globals)
		if stdinIsTerminal {
			fmt.Println()
		}
	default:
		log.Print("want at most one Starlark file name")
		return 1
	}

	// Print the global environment.
	if *showenv {
		for _, name := range globals.Keys() {
			if !strings.HasPrefix(name, "_") {
				fmt.Fprintf(os.Stderr, "%s = %s\n", name, globals[name])
			}
		}
	}

	return 0
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func noPredeclared(name string) bool { return false }

// A loadChecker typechecks a file and, recursively, the modules it
// loads, feeding each dependency's Interface into its dependents so
// that load()ed symbols have precise types instead of Any. Module
// names are interpreted as file names, exactly like the executor's
// load implementation (repl.MakeLoadOptions).
type loadChecker struct {
	opts   *syntax.FileOptions
	env    typecheck.Env
	cache  map[string]*typecheck.Interface // nil entry = check in progress
	errors []typecheck.Error               // accumulated across all modules
}

// checkFile typechecks a parsed, resolved file, checking its load
// dependencies first.
func (lc *loadChecker) checkFile(f *syntax.File) (*typecheck.Interface, error) {
	loads := make(map[string]*typecheck.Interface)
	for _, stmt := range f.Stmts {
		if load, ok := stmt.(*syntax.LoadStmt); ok {
			module := load.ModuleName()
			iface, err := lc.interfaceOf(module)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", load.Module.TokenPos, err)
			}
			loads[module] = iface
		}
	}
	res, err := typecheck.Check(f, lc.env, loads)
	if err != nil {
		return nil, err
	}
	lc.errors = append(lc.errors, res.Errors...)
	return res.Interface, nil
}

// interfaceOf returns the Interface of a loaded module, checking it
// (and its own dependencies) on first use.
func (lc *loadChecker) interfaceOf(module string) (*typecheck.Interface, error) {
	if iface, ok := lc.cache[module]; ok {
		if iface == nil {
			return nil, fmt.Errorf("cycle in load graph at %s", module)
		}
		return iface, nil
	}
	lc.cache[module] = nil // mark in progress
	f, _, err := starlark.SourceProgramOptions(lc.opts, module, nil, noPredeclared)
	if err != nil {
		return nil, err
	}
	iface, err := lc.checkFile(f)
	if err != nil {
		return nil, err
	}
	lc.cache[module] = iface
	return iface, nil
}
