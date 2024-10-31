package lua

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/nais/tester/lua/reporter"
	"github.com/nais/tester/lua/runner"
	"github.com/nais/tester/lua/spec"
	lua "github.com/yuin/gopher-lua"
)

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

	L.SetContext(ctx)

	L.Register("Save", spec.Save)
	L.Register("Ignore", spec.Ignore)
	L.Register("NotNull", spec.NotNull)

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
				helperFuncs[f.Name] = func(l *lua.LState) int {
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
				}
			}
		}
	}

	helperMod := L.SetFuncs(L.NewTable(), helperFuncs)
	L.SetGlobal("Helper", helperMod)

	if err := L.DoFile(filename); err != nil {
		s.reporter.Error(err.Error())
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

		s.reporter.RunTest(L.Context(), actualRunner.Name(), name, func(r reporter.Reporter) {
			ctx := runner.WithSaveFunc(L.Context(), s.save)
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
				r.Error(err.Error())
			}
		})
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

	var err error
	s.runners, s.cleanup, err = s.mgr.doSetup(L.Context(), s.cfg)
	if err != nil {
		L.RaiseError("error during setup: %v", err)
	}

	s.setupDone = true
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
