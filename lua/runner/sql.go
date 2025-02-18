package runner

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nais/tester/lua/spec"
	lua "github.com/yuin/gopher-lua"
)

var _ spec.Runner = (*SQL)(nil)

type SQL struct {
	db      *pgxpool.Pool
	results any
}

func NewSQLRunner(db *pgxpool.Pool) *SQL {
	return &SQL{db: db}
}

func (s *SQL) Name() string {
	return "sql"
}

func (s *SQL) HelperFunctions() []*spec.Function {
	defaultArgs := []spec.Argument{
		{Name: "query", Type: []spec.ArgumentType{spec.ArgumentTypeString}, Doc: "SQL query to execute"},
		{Name: "...", Type: []spec.ArgumentType{spec.ArgumentTypeString, spec.ArgumentTypeBoolean, spec.ArgumentTypeNumber}, Doc: "Arguments to the query"},
	}

	return []*spec.Function{
		{
			Name: "SQLExec",
			Args: defaultArgs,
			Doc:  "Execute some SQL. Will error if the SQL fails",
			Func: s.execHelper,
		},
		{
			Name:    "SQLQueryRow",
			Args:    defaultArgs,
			Doc:     "Execute some SQL. Returns a single row. Error if no rows returned",
			Func:    s.queryRowHelper,
			Returns: []spec.ArgumentType{spec.ArgumentTypeTable},
		},
		{
			Name:    "SQLQuery",
			Args:    defaultArgs,
			Doc:     "Execute some SQL. Will return multiple rows.",
			Func:    s.queryHelper,
			Returns: []spec.ArgumentType{spec.ArgumentTypeTable},
		},
	}
}

func (s *SQL) Functions() []*spec.Function {
	return []*spec.Function{
		{
			Name: "query",
			Args: []spec.Argument{
				{
					Name: "query",
					Type: []spec.ArgumentType{spec.ArgumentTypeString},
					Doc:  "The query to run",
				},
				{
					Name: "...",
					Type: []spec.ArgumentType{spec.ArgumentTypeString, spec.ArgumentTypeBoolean, spec.ArgumentTypeNumber},
					Doc:  "The query arguments",
				},
			},
			Doc:  "Query for multiple rows",
			Func: s.query,
		},
		{
			Name: "queryRow",
			Args: []spec.Argument{
				{
					Name: "query",
					Type: []spec.ArgumentType{spec.ArgumentTypeString},
					Doc:  "The query to run",
				},
				{
					Name: "...",
					Type: []spec.ArgumentType{spec.ArgumentTypeString, spec.ArgumentTypeBoolean, spec.ArgumentTypeNumber},
					Doc:  "The query arguments",
				},
			},
			Doc:  "Query for a single row. Will error if no rows returned",
			Func: s.queryRow,
		},
		StdCheckDefinition(s.check),
	}
}

func (s *SQL) check(L *lua.LState) int {
	tbl := L.CheckTable(1)

	StdCheck(L, tbl, s.results)
	return 0
}

// func (s *SQL) Run(ctx context.Context, logf func(format string, args ...any), body []byte, state map[string]any) error {
// 	f, err := parser.Parse(body, state)
// 	if err != nil {
// 		return fmt.Errorf("gql.Parse: %w", err)
// 	}

// 	return f.Execute(state, func() (any, error) {
// 		if rowTypeRegexp.MatchString(f.Query) {
// 			return s.queryRow(ctx, f)
// 		}
// 		return s.query(ctx, f)
// 	})
// }

func (s *SQL) query(L *lua.LState) int {
	res := s.doQuery(L)

	ret := make([]any, len(res))
	for i, row := range res {
		ret[i] = row
	}

	s.results = ret

	return 0
}

func (s *SQL) queryRow(L *lua.LState) int {
	s.results = s.doQueryRow(L)

	return 0
}

func (s *SQL) queryHelper(L *lua.LState) int {
	res := s.doQuery(L)

	if len(res) == 0 {
		L.Push(lua.LNil)
	} else {
		tbl := L.NewTable()
		for i, row := range res {
			r := L.NewTable()
			for k, v := range row {
				L.SetField(r, k, toLuaType(v))
			}
			L.SetTable(tbl, lua.LNumber(i+1), r)
		}
		L.Push(tbl)
	}

	return 1
}

func (s *SQL) queryRowHelper(L *lua.LState) int {
	res := s.doQueryRow(L)

	if len(res) == 0 {
		L.Push(lua.LNil)
	} else {
		tbl := L.NewTable()
		for k, v := range res {
			L.SetField(tbl, k, toLuaType(v))
		}
		L.Push(tbl)
	}

	return 1
}

func (s *SQL) execHelper(L *lua.LState) int {
	query := L.CheckString(1)
	args := vargs(L)

	ctx := L.Context()
	err := s.db.AcquireFunc(ctx, func(c *pgxpool.Conn) error {
		_, err := c.Exec(ctx, query, args...)
		return err
	})
	if err != nil {
		panic(fmt.Sprintf("sql.Run: unable to run query: %v", err))
	}

	return 0
}

func (s *SQL) doQuery(L *lua.LState) []map[string]any {
	query := L.CheckString(1)
	args := vargs(L)

	ctx := L.Context()
	ret := []map[string]any{}
	err := s.db.AcquireFunc(ctx, func(c *pgxpool.Conn) error {
		rows, err := c.Query(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("sql.Run: unable to execute query: %w", err)
		}
		defer rows.Close()
		ret, err = pgx.CollectRows(rows, pgx.RowToMap)
		if err != nil {
			return fmt.Errorf("sql.Run: unable to collect rows: %w", err)
		}

		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("sql.Run: unable to run query: %v", err))
	}

	return ret
}

func (s *SQL) doQueryRow(L *lua.LState) map[string]any {
	query := L.CheckString(1)
	args := vargs(L)

	ctx := L.Context()
	var ret map[string]any
	err := s.db.AcquireFunc(ctx, func(c *pgxpool.Conn) error {
		rows, err := c.Query(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("sql.Run: unable to execute query: %w", err)
		}
		defer rows.Close()
		ret, err = pgx.CollectOneRow(rows, pgx.RowToMap)
		return err
	})
	if err != nil {
		panic(fmt.Sprintf("sql.Run: unable to run query: %v", err))
	}

	return ret
}

func vargs(L *lua.LState) []any {
	args := []any{}
	for i := 2; i <= L.GetTop(); i++ {
		var v any
		vl := L.Get(i)
		switch vl.Type() {
		case lua.LTNumber:
			v = float64(vl.(lua.LNumber))
		case lua.LTString:
			v = string(vl.(lua.LString))
		case lua.LTBool:
			v = bool(vl.(lua.LBool))
		default:
			panic(fmt.Sprintf("vargs: unsupported type: %v", vl.Type()))
		}
		args = append(args, v)
	}
	return args
}

func toLuaType(v any) lua.LValue {
	switch vl := v.(type) {
	case int:
		return lua.LNumber(vl)
	case int64:
		return lua.LNumber(vl)
	case int32:
		return lua.LNumber(vl)
	case float64:
		return lua.LNumber(vl)
	case string:
		return lua.LString(vl)
	case bool:
		return lua.LBool(vl)
	default:
		if s, ok := v.(fmt.Stringer); ok {
			return lua.LString(s.String())
		}
		panic(fmt.Sprintf("toLuaType: unsupported type: %T", v))
	}
}
