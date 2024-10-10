package spec

import (
	lua "github.com/yuin/gopher-lua"
)

type Null struct{}

type SaveData struct {
	Name       string
	IgnoreNull bool
}

func Save(L *lua.LState) int {
	name := L.CheckString(1)
	ignoreNull := L.OptBool(2, false)

	ud := L.NewUserData()
	ud.Value = SaveData{
		Name:       name,
		IgnoreNull: ignoreNull,
	}

	L.Push(ud)

	return 1
}

type IgnoreData struct {
	IgnoreNull bool
}

func Ignore(L *lua.LState) int {
	ignoreNull := L.OptBool(1, false)

	ud := L.NewUserData()
	ud.Value = IgnoreData{
		IgnoreNull: ignoreNull,
	}

	L.Push(ud)

	return 1
}
