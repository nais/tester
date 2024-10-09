package spec

import (
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

type Argument struct {
	// Name of the argument
	// If the name ends with a ?, it is optional
	// If the name is ... it is variadic
	Name string
	Type []ArgumentType
	Doc  string
}

type Function struct {
	Name    string
	Args    []Argument
	Doc     string
	Func    lua.LGFunction
	Returns []ArgumentType
}
