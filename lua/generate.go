package lua

import (
	"fmt"
	"io"
	"reflect"
	"slices"
	"strings"

	"github.com/nais/tester/lua/spec"
	lua "github.com/yuin/gopher-lua"
)

const specFilename = "zz_spec.lua"

const base = `
-- This file is generated. Do not edit.


--- Ignore the field regardless of its value
---@return userdata
function Ignore()
  print("Ignore")
  ---@diagnostic disable-next-line: return-type-mismatch
  return {}
end

--- Ensure the field is not null, but allow any other value
---@return userdata
function NotNull()
	print("NotNull")
  ---@diagnostic disable-next-line: return-type-mismatch
	return {}
end

--- Save the field to the state. By default it will error if the field is null
---@param name string Name of the field in the state
---@param allowNull? boolean
---@return userdata
function Save(name, allowNull)
  print("Save: ", name, allowNull)
  ---@diagnostic disable-next-line: return-type-mismatch
  return {}
end

--- State variables
---@type table<string, any>
State = {}

--- Null ensures the value is null
---@type userdata
---@diagnostic disable-next-line: assign-type-mismatch
Null = {}

`

func GenerateSpec(w io.Writer, runners []spec.Runner, cfg any) {
	sb := &strings.Builder{}
	sb.WriteString(base)

	for _, r := range runners {
		specForRunner(sb, r)
	}

	sb.WriteString("--- Test case\n---@class Test\n")
	for _, r := range runners {
		sb.WriteString("---@field " + r.Name() + " fun(name: string, fn: fun(t: TestFunctionT" + r.Name() + "))\n")
	}

	sb.WriteString("Test = {}")

	helpers, err := combineHelpers(runners)
	if err != nil {
		panic(err)
	}

	if len(helpers) > 0 {
		sb.WriteString("\n\n--- Helper functions\n")
		sb.WriteString("---@class Helper\n")
		sb.WriteString("Helper = {}\n\n")
		for _, f := range helpers {
			writeFunc(sb, "Helper", f)
		}
	}

	writeConfig(sb, cfg)

	results := strings.TrimSpace(sb.String()) + "\n"
	_, _ = w.Write([]byte(results))
}

func specForRunner(sb *strings.Builder, r spec.Runner) {
	scope := "TestFunctionT" + r.Name()
	sb.WriteString("---@class " + scope + "\n")
	sb.WriteString("local " + scope + " = {}\n\n")

	for _, f := range r.Functions() {
		writeFunc(sb, scope, f)
	}
}

func writeFunc(sb *strings.Builder, scope string, f *spec.Function) {
	sb.WriteString("--- " + f.Doc + "\n")
	for _, a := range f.Args {
		sb.WriteString("---@param " + a.Name + " ")
		for i, t := range a.Type {
			if i != 0 {
				sb.WriteString("|")
			}

			sb.WriteString(t.String())
		}
		sb.WriteString("\n")
	}

	if len(f.Returns) > 0 {
		sb.WriteString("---@return ")
		for i, t := range f.Returns {
			if i != 0 {
				sb.WriteString("|")
			}

			sb.WriteString(t.String())
		}
		sb.WriteString("\n")
	}

	sb.WriteString("function " + scope + "." + f.Name + "(")
	for i, a := range f.Args {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(strings.TrimRight(a.Name, "?"))
	}
	sb.WriteString(")\n")
	sb.WriteString("  print(\"" + f.Name + "\")\n")

	if len(f.Returns) > 0 {
		sb.WriteString("  return ")
		switch f.Returns[0] {
		case spec.ArgumentTypeString:
			sb.WriteString("\"\"")
		case spec.ArgumentTypeNumber:
			sb.WriteString("0")
		case spec.ArgumentTypeBoolean:
			sb.WriteString("false")
		case spec.ArgumentTypeTable:
			sb.WriteString("{}")
		}

		sb.WriteString("\n")
	}

	sb.WriteString("end\n\n")
}

func combineHelpers(r []spec.Runner) ([]*spec.Function, error) {
	var funcs []*spec.Function
	for _, runner := range r {
		if h, ok := runner.(spec.HasHelperFunctions); ok {
			funcs = append(funcs, h.HelperFunctions()...)
		}
	}

	slices.SortFunc(funcs, func(a, b *spec.Function) int {
		return strings.Compare(a.Name, b.Name)
	})

	compact := slices.CompactFunc(funcs, func(a, b *spec.Function) bool {
		return a.Name == b.Name
	})

	if len(compact) != len(funcs) {
		return nil, fmt.Errorf("duplicate function names")
	}

	return funcs, nil
}

func ConvertToGoType(v lua.LValue) any {
	switch v.Type() {
	case lua.LTNil:
		return nil
	case lua.LTBool:
		return lua.LVAsBool(v)
	case lua.LTNumber:
		return lua.LVAsNumber(v)
	case lua.LTString:
		return lua.LVAsString(v)
	case lua.LTTable:
		tbl := v.(*lua.LTable)
		if tbl.Len() == 0 {
			// Treat as map
			m := make(map[any]any, tbl.Len())
			tbl.ForEach(func(k, v lua.LValue) {
				m[ConvertToGoType(k)] = ConvertToGoType(v)
			})

			return m
		} else {
			// Treat as list
			l := make([]any, 0, tbl.Len())
			tbl.ForEach(func(_, v lua.LValue) {
				l = append(l, ConvertToGoType(v))
			})

			return l
		}
	default:
		panic("unknown type" + v.Type().String())
	}
}

func writeConfig(sb *strings.Builder, cfg any) {
	if cfg == nil {
		return
	}

	t := reflect.TypeOf(cfg)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		panic("config must be a pointer to a struct")
	}

	fields := t.Elem().NumField()
	if fields == 0 {
		return
	}

	sb.WriteString("--- Configuration\n")
	sb.WriteString("---@class Config\n")
	for i := 0; i < fields; i++ {
		field := t.Elem().Field(i)
		sb.WriteString("---@field " + field.Name + " ")
		switch field.Type.Kind() {
		case reflect.String:
			sb.WriteString("string")
		case reflect.Int:
			sb.WriteString("number")
		case reflect.Bool:
			sb.WriteString("boolean")
		default:
			panic(fmt.Sprintf("unknown type: %s", field.Type.Kind()))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("Config = {\n")

	v := reflect.ValueOf(cfg).Elem()
	for i := 0; i < fields; i++ {
		field := t.Elem().Field(i)
		sb.WriteString("  " + field.Name + " = ")
		switch field.Type.Kind() {
		case reflect.String:
			fmt.Fprintf(sb, "%q", v.Field(i).String())
		case reflect.Int:
			fmt.Fprintf(sb, "%d", v.Field(i).Int())
		case reflect.Bool:
			fmt.Fprintf(sb, "%t", v.Field(i).Bool())
		default:
			panic(fmt.Sprintf("unknown type: %s", field.Type.Kind()))
		}
		sb.WriteString(",\n")
	}
	sb.WriteString("}\n")
}
