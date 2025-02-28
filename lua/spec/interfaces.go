package spec

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

type Runner interface {
	Name() string
	Functions() []*Function
}

type HasHelperFunctions interface {
	HelperFunctions() []*Function
}

type RunnerAfterTest interface {
	AfterTest(ctx context.Context)
}

type StringEnum []string

func (e StringEnum) String() string {
	var v []string
	for _, s := range e {
		v = append(v, strconv.Quote(s))
	}
	return strings.Join(v, " | ")
}

type ArgumentType interface {
	String() string
}

type ArgumentTypeA int

func (a ArgumentTypeA) String() string {
	switch a {
	case ArgumentTypeString:
		return "string"
	case ArgumentTypeNumber:
		return "number"
	case ArgumentTypeBoolean:
		return "boolean"
	case ArgumentTypeTable:
		return "table"
	default:
		panic(fmt.Sprintf("unknown type: %d", a))
	}
}

const (
	ArgumentTypeString ArgumentTypeA = iota
	ArgumentTypeNumber
	ArgumentTypeBoolean
	ArgumentTypeTable
)

type ArgumentTypeMetatable string

func (a ArgumentTypeMetatable) String() string {
	return string(a)
}

type Argument struct {
	// Name of the argument
	// If the name ends with a ?, it is optional
	// If the name is ... it is variadic
	Name string
	Type []ArgumentType
	Doc  string
}

func (a Argument) String() string {
	types := make([]string, len(a.Type))
	for i, t := range a.Type {
		types[i] = t.String()
	}
	return fmt.Sprintf("%s: %s", a.Name, strings.Join(types, "|"))
}

type Function struct {
	Name     string
	Args     []Argument
	Doc      string
	Func     lua.LGFunction
	Returns  []ArgumentType
	method   bool
	overload *Function
}

func (f Function) WriteOverloadTo(sb *strings.Builder, scope string) {
	args := []string{
		"self: " + scope,
	}

	for _, a := range f.Args {
		args = append(args, a.String())
	}

	var returns []string
	for _, r := range f.Returns {
		returns = append(returns, r.String())
	}

	ret := fmt.Sprintf("fun(%s)", strings.Join(args, ", "))
	if len(returns) > 0 {
		ret += ": " + strings.Join(returns, ", ")
	}
	sb.WriteString("---@overload ")
	sb.WriteString(ret + "\n")
}

func (f Function) WriteTo(sb *strings.Builder, scope string) {
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

	if f.overload != nil {
		f.overload.WriteOverloadTo(sb, scope)
	}

	sb.WriteString("function " + scope)
	if f.method {
		sb.WriteString(":")
	} else {
		sb.WriteString(".")
	}
	sb.WriteString(f.Name + "(")
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
		case ArgumentTypeString:
			sb.WriteString("\"\"")
		case ArgumentTypeNumber:
			sb.WriteString("0")
		case ArgumentTypeBoolean:
			sb.WriteString("false")
		case ArgumentTypeTable:
			sb.WriteString("{}")
		default:
			switch f.Returns[0].(type) {
			case ArgumentTypeTableLiteral, ArgumentTypeArray, ArgumentTypeMetatable:
				sb.WriteString("{}")
			}
		}

		sb.WriteString("\n")
	}

	sb.WriteString("end\n\n")
}

type ArgumentTypeArray struct {
	Type ArgumentType
}

func (a ArgumentTypeArray) String() string {
	return fmt.Sprintf("%s[]", a.Type)
}

type ArgumentTypeTableLiteralField struct {
	Name string
	Type ArgumentType
}

type ArgumentTypeTableLiteral struct {
	Fields []ArgumentTypeTableLiteralField
}

func (o ArgumentTypeTableLiteral) String() string {
	var fields []string
	for _, f := range o.Fields {
		fields = append(fields, fmt.Sprintf("%s: %s", f.Name, f.Type))
	}
	return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
}

type TypemetatableFunction struct {
	Function
}

type TypemetatableGetSet struct {
	Name         string
	Doc          string
	Func         lua.LGFunction
	SetArguments []Argument
	GetReturns   []ArgumentType
}

type Typemetatable struct {
	Name    string
	Init    *Function
	GetSet  []TypemetatableGetSet
	Methods []Function
}

func (t Typemetatable) String() string {
	sb := strings.Builder{}
	sb.WriteString("---@class ")
	sb.WriteString(t.Name)
	sb.WriteString("\n")
	sb.WriteString(t.Name)
	sb.WriteString(" = {}\n")

	if t.Init != nil {
		t.Init.Name = "new"
		t.Init.Returns = []ArgumentType{ArgumentTypeMetatable(t.Name)}
		t.Init.WriteTo(&sb, t.Name)
	}

	for _, i := range t.GetSet {
		if len(i.GetReturns) == 0 && len(i.SetArguments) == 0 {
			continue
		}

		var setFunc *Function
		if len(i.SetArguments) > 0 {
			setFunc = &Function{
				Name:    i.Name,
				Doc:     i.Doc,
				Args:    i.SetArguments,
				Returns: nil,
				Func:    i.Func,
				method:  true,
			}
		}

		if len(i.GetReturns) > 0 {
			Function{
				Name:     i.Name,
				Doc:      i.Doc,
				Returns:  i.GetReturns,
				Func:     i.Func,
				method:   true,
				overload: setFunc,
			}.WriteTo(&sb, t.Name)
		} else {
			setFunc.WriteTo(&sb, t.Name)
		}
	}

	for _, m := range t.Methods {
		m.method = true
		m.WriteTo(&sb, t.Name)
	}

	return sb.String()
}
