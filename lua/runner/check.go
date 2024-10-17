package runner

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/nais/tester/lua/spec"
	lua "github.com/yuin/gopher-lua"
)

type (
	contextKey int
	SaveFunc   func(key string, value any)
)

const (
	ctxSaveFunc contextKey = iota
)

const (
	emptyListOrMapString        = "[[[ empty_list_or_map ]]]"
	containsString              = "[[[ contains ]]]"
	containsStringCaseSensitive = "[[[ contains_case_sensitive ]]]"
)

type Reporter interface {
	Error(msg string, args ...any)
}

func WithSaveFunc(ctx context.Context, fn SaveFunc) context.Context {
	return context.WithValue(ctx, ctxSaveFunc, fn)
}

func StdCheck(L *lua.LState, tbl *lua.LTable, b any) {
	if err := StdCheckError(L.Context(), tbl, b); err != nil {
		L.RaiseError(err.Error())
	}
}

func StdCheckError(ctx context.Context, tbl *lua.LTable, b any) error {
	toSave := make(map[string]string)
	a, opts := convertToCheck("", tbl, toSave, nil)

	diff := cmp.Diff(a, b, opts...)
	if diff != "" {
		return fmt.Errorf("diff -want +got:\n%v", diff)
	}

	saveFunc := ctx.Value(ctxSaveFunc).(SaveFunc)

	for path, name := range toSave {
		appendStore(name, path, b, saveFunc)
	}
	return nil
}

func appendStore(key, path string, body any, fn SaveFunc) {
	if _, ok := body.(map[string]any); !ok {
		return
	}

	var (
		root = body.(map[string]any)
		val  interface{}
	)

	pathParts := strings.Split(path, ".")

	for i := 0; i < len(pathParts); i++ {
		if i == len(pathParts)-1 {
			// Last element of pathParts
			val = root[pathParts[i]]
			break
		} else if i < len(pathParts)-2 {
			// check if next value is a list
			intVal, err := strconv.Atoi(pathParts[i+1])
			if err == nil {
				slice := root[pathParts[i]].([]interface{})
				root = slice[intVal].(map[string]interface{})
				// skip one iteration on lists
				i += 1
				continue
			}
		}
		mp, ok := root[pathParts[i]].(map[string]interface{})
		if ok {
			root = mp
		}
	}
	fn(key, val)
}

func StdCheckDefinition(fn lua.LGFunction) *spec.Function {
	return &spec.Function{
		Name: "check",
		Args: []spec.Argument{
			{
				Name: "resp",
				Type: []spec.ArgumentType{spec.ArgumentTypeTable},
				Doc:  "The response to check",
			},
		},
		Doc:  "Check comment",
		Func: fn,
	}
}

// convertToCheck converts a lua table to a go map suitable for comparison
// Path is the path to the current value in the table, should be empty on first call
// v is the current value in the table
// toSave is a map of paths to names of values that should be saved
// opts is a list of cmp options
// It returns the converted value and the updated list of cmp options
func convertToCheck(path string, v lua.LValue, toSave map[string]string, opts cmp.Options) (any, cmp.Options) {
	switch v.Type() {
	case lua.LTNil:
		return nil, opts
	case lua.LTBool:
		return bool(lua.LVAsBool(v)), opts
	case lua.LTNumber:
		return float64(lua.LVAsNumber(v)), opts
	case lua.LTString:
		return string(lua.LVAsString(v)), opts
	case lua.LTTable:
		tbl := v.(*lua.LTable)
		if tbl.Len() == 0 {
			// Treat as map
			m := make(map[string]any, tbl.Len())
			tbl.ForEach(func(k, v lua.LValue) {
				key := lua.LVAsString(k)
				var val any
				val, opts = convertToCheck(path+"."+key, v, toSave, opts)
				m[key] = val
			})

			if len(m) == 0 {
				// As Lua tables can be both lists and maps, we cannot distinguish between an empty list and an empty map
				// So we treat an empty table as an OR between an empty list and an empty map
				return emptyListOrMapString, append(opts, emptyContainer(path))
			}

			return m, opts
		} else {
			// Treat as list
			l := make([]any, 0, tbl.Len())
			tbl.ForEach(func(i, v lua.LValue) {
				// Lua indexes are 1-based
				luaIdx := lua.LVAsNumber(i)
				goIdx := int(luaIdx) - 1
				var val any
				val, opts = convertToCheck(path+"."+strconv.Itoa(goIdx), v, toSave, opts)
				l = append(l, val)
			})

			return l, opts
		}
	case lua.LTUserData:
		ud := v.(*lua.LUserData)
		// We return a string to indicate which function is used in the Lua code

		switch v := ud.Value.(type) {
		case spec.SaveData:
			toSave[path] = v.Name
			if v.AllowNull {
				return "[[[ save_allow_null ]]]", append(opts, allowNull(path))
			}
			return "[[[ save ]]]", append(opts, notNull(path))
		case spec.IgnoreData:
			return "[[[ ignore ]]]", append(opts, allowNull(path))
		case spec.NotNullData:
			return "[[[ not_null ]]]", append(opts, notNull(path))
		case spec.Null:
			return nil, opts
		case spec.ContainsString:
			value := containsString
			if v.CaseSensitive {
				value = containsStringCaseSensitive
			}
			return value, append(opts, stringContains(path, v.Contains, v.CaseSensitive))
		default:
			panic("unknown userdata type " + fmt.Sprintf("%T", v))
		}

	default:
		panic("unknown type" + v.Type().String())
	}
}

// notNull ensures that the value is not null
func notNull(path string) cmp.Option {
	return cmp.FilterPath(ignorePath(path), cmp.Comparer(cmpNotNull))
}

// allowNull allows the value to be null and will just ignore the value
func allowNull(path string) cmp.Option {
	return cmp.FilterPath(ignorePath(path), cmp.Ignore())
}

// stringContains ensures the string contains a substring
func stringContains(path, substr string, caseSensitive bool) cmp.Option {
	return cmp.FilterPath(ignorePath(path), cmp.Comparer(contains(substr, caseSensitive)))
}

func contains(substr string, caseSensitive bool) func(a, b string) bool {
	return func(a, b string) bool {
		var toCheck string
		if a == containsString || a == containsStringCaseSensitive {
			toCheck = b
		} else if b == containsString || b == containsStringCaseSensitive {
			toCheck = a
		} else {
			return false
		}

		if caseSensitive {
			return strings.Contains(toCheck, substr)
		}
		return strings.Contains(
			strings.ToLower(toCheck), strings.ToLower(substr),
		)
	}
}

func cmpNotNull(a, b any) bool {
	if a == nil || b == nil {
		return false
	}
	return true
}

func ignorePath(path string) func(p cmp.Path) bool {
	return func(p cmp.Path) bool {
		s := ""
		for _, pe := range p {
			switch pe := pe.(type) {
			case cmp.MapIndex:
				s += "." + pe.Key().String()
			case cmp.SliceIndex:
				s += "." + strconv.Itoa(pe.Key())
			}
		}
		return s == path
	}
}

func emptyContainer(path string) cmp.Option {
	return cmp.FilterPath(ignorePath(path), cmp.Comparer(cmpEmptyContainer))
}

func cmpEmptyContainer(a, b any) bool {
	var toCheck any
	if s, ok := a.(string); ok && s == emptyListOrMapString {
		toCheck = b
	} else if s, ok := b.(string); ok && s == emptyListOrMapString {
		toCheck = a
	} else {
		return false
	}

	switch v := toCheck.(type) {
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	}
	return false
}
