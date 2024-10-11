package spec

import (
	lua "github.com/yuin/gopher-lua"
)

type Null struct{}

type SaveData struct {
	Name      string
	AllowNull bool
}

func Save(L *lua.LState) int {
	name := L.CheckString(1)
	failOnNull := L.OptBool(2, false)

	ud := L.NewUserData()
	ud.Value = SaveData{
		Name:      name,
		AllowNull: failOnNull,
	}

	L.Push(ud)

	return 1
}

type IgnoreData struct{}

func Ignore(L *lua.LState) int {
	ud := L.NewUserData()
	ud.Value = IgnoreData{}

	L.Push(ud)

	return 1
}

type NotNullData struct{}

func NotNull(L *lua.LState) int {
	ud := L.NewUserData()
	ud.Value = NotNullData{}

	L.Push(ud)

	return 1
}
