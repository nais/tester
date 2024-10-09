package lua

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nais/tester/lua/spec"
	lua "github.com/yuin/gopher-lua"
)

type config struct {
	Field     string
	Other     int
	Supported bool
}

func TestGenerate(t *testing.T) {
	buf := &bytes.Buffer{}
	GenerateSpec(buf, []spec.Runner{
		&GQLRunner{},
		&RESTRunner{},
	}, &config{Supported: true, Other: 42, Field: "test"})

	expected := `-- This file is generated. Do not edit.


--- Ignore the field, if notNull is true, it will ensure the field is not null
---@param notNull? boolean
---@return string
function Ignore(notNull)
  print("Ignore: ", notNull)
  return ""
end

--- Save the field to the state. By default it will error if the field is null
---@param name string Name of the field in the state
---@param ignoreNull? boolean
---@return string
function Save(name, ignoreNull)
  print("Save: ", name, ignoreNull)
  return ""
end

--- State variables
---@type table<string, any>
State = {}

---@class TestFunctionTgql
local TestFunctionTgql = {}

--- Query comment
---@param query string
function TestFunctionTgql.query(query)
  print("query")
end

--- Check comment
---@param resp table|string
function TestFunctionTgql.check(resp)
  print("check")
end

---@class TestFunctionTrest
local TestFunctionTrest = {}

--- Send request
---@param method string
---@param path string
---@param body string
function TestFunctionTrest.send(method, path, body)
  print("send")
end

--- Check comment
---@param statusCode number
---@param resp table
function TestFunctionTrest.check(statusCode, resp)
  print("check")
end

--- Test case
---@class Test
---@field gql fun(name: string, fn: fun(t: TestFunctionTgql))
---@field rest fun(name: string, fn: fun(t: TestFunctionTrest))
Test = {}

--- Helper functions
---@class Helper
Helper = {}

--- Execute some SQL. Will error if the SQL fails
---@param query string
function Helper.SQLExec(query)
  print("SQLExec")
end

--- Execute some SQL. Will return multiple rows.
---@param query string
---@return table
function Helper.SQLQuery(query)
  print("SQLQuery")
  return {}
end

--- Execute some SQL. Returns a single row. Error if no rows returned
---@param query string
---@return table|boolean|number|string
function Helper.SQLQueryRow(query)
  print("SQLQueryRow")
  return {}
end

--- Configuration
---@class Config
---@field Field string
---@field Other number
---@field Supported boolean
Config = {
  Field = "test",
  Other = 42,
  Supported = true,
}
`

	if diff := cmp.Diff(buf.String(), expected); diff != "" {
		t.Errorf("Generate() mismatch (-want +got):\n%s", diff)
	}
}

// Some comment about the runner
type GQLRunner struct{}

func (r *GQLRunner) Name() string {
	return "gql"
}

func (r *GQLRunner) Functions() []*spec.Function {
	return []*spec.Function{
		{
			Name: "query",
			Args: []spec.Argument{
				{
					Name: "query",
					Type: []spec.ArgumentType{spec.ArgumentTypeString},
					Doc:  "The query to run",
				},
			},
			Doc:  "Query comment",
			Func: r.query,
		},
		{
			Name: "check",
			Args: []spec.Argument{
				{
					Name: "resp",
					Type: []spec.ArgumentType{spec.ArgumentTypeTable, spec.ArgumentTypeString},
					Doc:  "The response to check",
				},
			},
			Doc:  "Check comment",
			Func: r.check,
		},
	}
}

func (r *GQLRunner) query(L *lua.LState) int {
	query := L.CheckString(1)
	r.Query(query)
	return 0
}

// Query comment
func (r *GQLRunner) Query(query string) {}

func (r *GQLRunner) check(L *lua.LState) int {
	v := L.Get(1)
	var resp any
	switch v.Type() {
	case lua.LTTable:
		resp = ConvertToGoType(v)
	case lua.LTString:
		resp = lua.LVAsString(v)
	default:
		L.ArgError(1, "expected table or string")
		return 0
	}
	r.Check(resp)
	return 0
}

// Check comment
func (r *GQLRunner) Check(resp any) {}

type RESTRunner struct{}

func (r *RESTRunner) Functions() []*spec.Function {
	return []*spec.Function{
		{
			Name: "send",
			Args: []spec.Argument{
				{
					Name: "method",
					Type: []spec.ArgumentType{spec.ArgumentTypeString},
					Doc:  "HTTP Method",
				},
				{
					Name: "path",
					Type: []spec.ArgumentType{spec.ArgumentTypeString},
					Doc:  "HTTP path",
				},
				{
					Name: "body",
					Type: []spec.ArgumentType{spec.ArgumentTypeString},
					Doc:  "Request body",
				},
			},
			Doc:  "Send request",
			Func: r.query,
		},
		{
			Name: "check",
			Args: []spec.Argument{
				{
					Name: "statusCode",
					Type: []spec.ArgumentType{spec.ArgumentTypeNumber},
					Doc:  "Response status code",
				},
				{
					Name: "resp",
					Type: []spec.ArgumentType{spec.ArgumentTypeTable},
					Doc:  "The response to check",
				},
			},
			Doc:  "Check comment",
			Func: r.check,
		},
	}
}

func (r *RESTRunner) Name() string {
	return "rest"
}

func (r *RESTRunner) query(L *lua.LState) int {
	method := L.CheckString(1)
	path := L.CheckString(2)
	body := L.CheckString(3)
	r.Send(method, path, body)

	return 0
}

// Query comment
func (r *RESTRunner) Send(method string, path string, body string) {}

func (r *RESTRunner) check(L *lua.LState) int {
	statusCode := L.CheckInt(1)
	resp := L.CheckTable(2)

	r.Check(statusCode, ConvertToGoType(resp))
	return 0
}

// Check comment
func (r *RESTRunner) Check(statusCode int, resp any) {}

func (r *RESTRunner) HelperFunctions() []*spec.Function {
	return []*spec.Function{
		{
			Name: "SQLExec",
			Args: []spec.Argument{
				{Name: "query", Type: []spec.ArgumentType{spec.ArgumentTypeString}, Doc: "SQL query to execute"},
			},
			Doc:  "Execute some SQL. Will error if the SQL fails",
			Func: func(l *lua.LState) int { return 0 },
		},
		{
			Name: "SQLQueryRow",
			Args: []spec.Argument{
				{Name: "query", Type: []spec.ArgumentType{spec.ArgumentTypeString}, Doc: "SQL query to execute"},
			},
			Doc:     "Execute some SQL. Returns a single row. Error if no rows returned",
			Func:    func(l *lua.LState) int { return 0 },
			Returns: []spec.ArgumentType{spec.ArgumentTypeTable, spec.ArgumentTypeBoolean, spec.ArgumentTypeNumber, spec.ArgumentTypeString},
		},
		{
			Name: "SQLQuery",
			Args: []spec.Argument{
				{Name: "query", Type: []spec.ArgumentType{spec.ArgumentTypeString}, Doc: "SQL query to execute"},
			},
			Doc:     "Execute some SQL. Will return multiple rows.",
			Func:    func(l *lua.LState) int { return 0 },
			Returns: []spec.ArgumentType{spec.ArgumentTypeTable},
		},
	}
}
