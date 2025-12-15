package lua

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/nais/tester/lua/reporter"
	"github.com/nais/tester/lua/runner"
	"github.com/nais/tester/lua/spec"
	lua "github.com/yuin/gopher-lua"
)

// testReporter is a wrapper that holds the current test's reporter
type testReporter struct {
	reporter reporter.Reporter
}

type suite struct {
	// setupDone should be set to true after the setup function has been called. This should
	// be done when the first test is run or first helper is invoked.
	setupDone bool
	runners   []spec.Runner
	mgr       *Manager
	state     *lua.LTable
	reporter  reporter.Reporter
	cfg       any
	cleanup   func()
}

func newSuite(mgr *Manager, reporter reporter.Reporter) *suite {
	var cfg any
	if mgr.newConfigFn != nil {
		cfg = mgr.newConfigFn()
	}
	return &suite{
		mgr:      mgr,
		reporter: reporter,
		cfg:      cfg,
	}
}

func (s *suite) run(ctx context.Context, filename string) {
	L := lua.NewState()
	defer L.Close()

	defer func() {
		if s.cleanup != nil {
			s.cleanup()
		}
	}()

	// Set file-level reporter in context so top-level calls can log info
	ctx = runner.WithReporter(ctx, s.reporter)
	L.SetContext(ctx)

	L.Register("Save", spec.Save)
	L.Register("Ignore", spec.Ignore)
	L.Register("NotNull", spec.NotNull)
	L.Register("Contains", spec.Contains)

	nullD := L.NewUserData()
	nullD.Value = spec.Null{}
	L.SetGlobal("Null", nullD)

	ud := L.NewTable()
	L.SetGlobal("Config", ud)

	s.state = L.NewTable()
	L.SetGlobal("State", s.state)

	tests := map[string]lua.LGFunction{}

	for _, r := range s.mgr.runners {
		tests[r.Name()] = s.newTest(r.Name(), L)
	}

	mod := L.SetFuncs(L.NewTable(), tests)
	L.SetGlobal("Test", mod)

	helperFuncs := map[string]lua.LGFunction{}
	for _, r := range s.mgr.runners {
		if h, ok := r.(spec.HasHelperFunctions); ok {
			for _, f := range h.HelperFunctions() {
				helperFuncs[f.Name] = s.wrapHelper(f, func(l *lua.LState) int {
					s.setup(L)

					var actualRunner spec.Runner
					for _, i := range s.runners {
						if i.Name() == r.Name() {
							actualRunner = i
							break
						}
					}

					if actualRunner == nil {
						L.RaiseError("runner %q not found", r.Name())
					}

					var fn lua.LGFunction
					for _, inf := range actualRunner.(spec.HasHelperFunctions).HelperFunctions() {
						if inf.Name == f.Name {
							fn = inf.Func
							break
						}
					}

					if fn == nil {
						L.RaiseError("helper function %q not found", f.Name)
					}

					return fn(L)
				})
			}
		}
	}

	for _, f := range s.mgr.helpers {
		helperFuncs[f.Name] = s.wrapHelper(f, func(l *lua.LState) int {
			s.setup(L)
			return f.Func(l)
		})
	}

	helperMod := L.SetFuncs(L.NewTable(), helperFuncs)
	L.SetGlobal("Helper", helperMod)

	for _, t := range s.mgr.typeMetatable {
		mt := L.NewTypeMetatable(t.Name)
		L.SetGlobal(t.Name, mt)
		// static attributes
		L.SetField(mt, "new", L.NewFunction(s.wrapTypemetatable(t, "new", func(L *lua.LState) int {
			s.setup(L)
			return t.Init.Func(L)
		})))
		// methods
		index := map[string]lua.LGFunction{}
		for _, f := range t.GetSet {
			index[f.Name] = s.wrapTypemetatable(t, f.Name, func(L *lua.LState) int {
				s.setup(L)
				return f.Func(L)
			})
		}

		for _, f := range t.Methods {
			index[f.Name] = s.wrapTypemetatable(t, f.Name, func(L *lua.LState) int {
				s.setup(L)
				return f.Func(L)
			})
		}
		L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), index))

	}

	if err := L.DoFile(filename); err != nil {
		s.reporter.ReportError(reporter.NewError("%s", err.Error()))
	}
}

func (s *suite) newTest(runnerName string, _ *lua.LState) lua.LGFunction {
	return func(L *lua.LState) int {
		name := L.CheckString(1)
		fn := L.CheckFunction(2)

		s.setup(L)

		var actualRunner spec.Runner
		for _, r := range s.runners {
			if r.Name() == runnerName {
				actualRunner = r
				break
			}
		}

		if actualRunner == nil {
			L.RaiseError("runner %q not found", runnerName)
		}

		// Save the current context to restore after the test
		ctxBeforeTest := L.Context()

		s.reporter.RunTest(L.Context(), actualRunner.Name(), name, func(r reporter.Reporter) {
			ctx := runner.WithSaveFunc(L.Context(), s.save)
			ctx = runner.WithReporter(ctx, r)
			L.SetContext(ctx)

			mp := map[string]lua.LGFunction{}
			for _, f := range actualRunner.Functions() {
				mp[f.Name] = func(l *lua.LState) int {
					return f.Func(l)
				}
			}

			mod := L.SetFuncs(L.NewTable(), mp)

			err := L.CallByParam(lua.P{
				Fn:      fn,
				Protect: true,
			}, mod)
			if err != nil {
				// Check if this was a CheckError with structured data
				if checkErr, ok := runner.GetCheckError(L.Context()); ok {
					r.ReportError(reporter.NewDiffError(checkErr.Diff, checkErr.Expected, checkErr.Actual))
				} else {
					r.ReportError(reporter.NewError("%s", err.Error()))
				}
			}
		})

		// Restore the file-level context so subsequent top-level code uses the file reporter
		L.SetContext(ctxBeforeTest)

		if hook, ok := actualRunner.(spec.RunnerAfterTest); ok {
			hook.AfterTest(L.Context())
		}

		return 0
	}
}

func (s *suite) setup(L *lua.LState) {
	if s.setupDone {
		return
	}

	t := ConvertToGoType(L.GetGlobal("Config"))

	if err := mapstructure.Decode(t, s.cfg); err != nil {
		L.RaiseError("error decoding config: %v", err)
	}

	// Preserve the reporter from the current context
	currentReporter := runner.GetReporter(L.Context())

	var err error
	var ctx context.Context
	ctx, s.runners, s.cleanup, err = s.mgr.doSetup(L.Context(), s.cfg)
	if err != nil {
		L.RaiseError("error during setup: %v", err)
	}

	// Re-attach the reporter to the new context
	if currentReporter != nil {
		ctx = runner.WithReporter(ctx, currentReporter)
	}

	L.SetContext(ctx)

	s.setupDone = true
}

// wrapHelper wraps a helper function to automatically log info about its invocation
func (s *suite) wrapHelper(f *spec.Function, fn lua.LGFunction) lua.LGFunction {
	return func(L *lua.LState) int {
		// Collect arguments for logging with names from spec
		args := make([]reporter.InfoArg, 0, L.GetTop())
		for i := 1; i <= L.GetTop(); i++ {
			value := formatLuaValue(L.Get(i))
			// Get argument name if available
			var argName string
			if i-1 < len(f.Args) {
				argName = strings.TrimSuffix(f.Args[i-1].Name, "?")
				if argName == "..." {
					argName = "" // Variadic args have no name
				}
			}
			args = append(args, reporter.InfoArg{Name: argName, Value: value})
		}

		// Log the helper call
		runner.Info(L.Context(), reporter.Info{
			Type:  reporter.InfoTypeHelper,
			Title: fmt.Sprintf("Helper.%s", f.Name),
			Args:  args,
		})

		return fn(L)
	}
}

// wrapTypemetatable wraps a typemetatable method to automatically log info about its invocation
func (s *suite) wrapTypemetatable(t *spec.Typemetatable, methodName string, fn lua.LGFunction) lua.LGFunction {
	// Find the argument definitions for this method
	var argDefs []spec.Argument
	if methodName == "new" && t.Init != nil {
		argDefs = t.Init.Args
	} else {
		for _, m := range t.Methods {
			if m.Name == methodName {
				argDefs = m.Args
				break
			}
		}
		if argDefs == nil {
			for _, gs := range t.GetSet {
				if gs.Name == methodName {
					argDefs = gs.SetArguments
					break
				}
			}
		}
	}

	return func(L *lua.LState) int {
		// Collect arguments for logging (skip first arg for methods as it's self)
		startArg := 1
		if methodName != "new" {
			startArg = 2 // Skip self for instance methods
		}

		args := make([]reporter.InfoArg, 0, L.GetTop())
		for i := startArg; i <= L.GetTop(); i++ {
			value := formatLuaValue(L.Get(i))
			argIdx := i - startArg
			var argName string
			if argIdx < len(argDefs) {
				argName = strings.TrimSuffix(argDefs[argIdx].Name, "?")
				if argName == "..." {
					argName = "" // Variadic args have no name
				}
			}
			args = append(args, reporter.InfoArg{Name: argName, Value: value})
		}

		// Log the method call
		var title string
		if methodName == "new" {
			title = fmt.Sprintf("%s.new", t.Name)
		} else {
			title = fmt.Sprintf("%s:%s", t.Name, methodName)
		}

		runner.Info(L.Context(), reporter.Info{
			Type:  reporter.InfoTypeHelper,
			Title: title,
			Args:  args,
		})

		return fn(L)
	}
}

// formatLuaValue formats a Lua value for display in logs
func formatLuaValue(v lua.LValue) string {
	switch v.Type() {
	case lua.LTString:
		return lua.LVAsString(v)
	case lua.LTNumber:
		return fmt.Sprintf("%v", lua.LVAsNumber(v))
	case lua.LTBool:
		return fmt.Sprintf("%v", lua.LVAsBool(v))
	case lua.LTTable:
		return "{...}"
	case lua.LTNil:
		return "nil"
	default:
		return v.String()
	}
}

func (s *suite) save(key string, value any) {
	var val lua.LValue
	switch v := value.(type) {
	case string:
		val = lua.LString(v)
	case int:
		val = lua.LNumber(v)
	case float64:
		val = lua.LNumber(v)
	case bool:
		val = lua.LBool(v)
	default:
		panic(fmt.Sprintf("unsupported save type: %T", value))
	}

	s.state.RawSet(lua.LString(key), val)
}
