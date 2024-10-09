package lua

import (
	"context"
	"fmt"

	"github.com/nais/tester/lua/runner"
	"github.com/nais/tester/lua/spec"
	lua "github.com/yuin/gopher-lua"
)

type suite struct {
	// setupDone should be set to true after the setup function has been called. This should
	// be done when the first test is run or first helper is invoked.
	setupDone bool
	runners   []spec.Runner
	cfg       any
	state     *lua.LTable
	reporter  Reporter
}

func newSuite(cfg any, runners []spec.Runner, reporter Reporter) *suite {
	return &suite{
		cfg:      cfg,
		runners:  runners,
		reporter: reporter,
	}
}

func (s *suite) run(ctx context.Context, filename string) {
	L := lua.NewState()
	defer L.Close()

	L.SetContext(ctx)

	L.Register("Save", spec.Save)
	L.Register("Ignore", spec.Ignore)

	ud := L.NewUserData()
	ud.Value = s.cfg
	L.SetGlobal("Config", ud)

	s.state = L.NewTable()
	L.SetGlobal("State", s.state)

	tests := map[string]lua.LGFunction{}

	for _, r := range s.runners {
		tests[r.Name()] = s.newTest(r, L)
	}

	mod := L.SetFuncs(L.NewTable(), tests)
	L.SetGlobal("Test", mod)

	helperFuncs := map[string]lua.LGFunction{}
	for _, r := range s.runners {
		if h, ok := r.(spec.HasHelperFunctions); ok {
			for _, f := range h.HelperFunctions() {
				helperFuncs[f.Name] = func(l *lua.LState) int {
					s.setup()
					return f.Func(l)
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

func (s *suite) newTest(tech spec.Runner, _ *lua.LState) lua.LGFunction {
	return func(L *lua.LState) int {
		name := L.CheckString(1)
		fn := L.CheckFunction(2)

		s.reporter.RunTest(L.Context(), tech.Name(), name, func(r Reporter) {
			ctx := runner.WithSaveFunc(L.Context(), s.save)
			L.SetContext(ctx)

			s.setup()

			mp := map[string]lua.LGFunction{}
			for _, f := range tech.Functions() {
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

func (s *suite) setup() {
	if s.setupDone {
		return
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
