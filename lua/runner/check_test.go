package runner

import (
	"context"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestStdCheckErrorComparesEquivalentNumericTypes(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	expected := L.NewTable()
	L.SetField(expected, "integer", lua.LNumber(1))
	L.SetField(expected, "decimal", lua.LNumber(1.5))

	ctx := WithSaveFunc(context.Background(), func(string, any) {})
	err := StdCheckError(ctx, expected, map[string]any{
		"integer": int64(1),
		"decimal": float32(1.5),
	})
	if err != nil {
		t.Fatalf("StdCheckError() returned an error: %v", err)
	}
}

func TestStdCheckErrorRejectsDifferentNumericValues(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	expected := L.NewTable()
	L.SetField(expected, "value", lua.LNumber(1))

	ctx := WithSaveFunc(context.Background(), func(string, any) {})
	if err := StdCheckError(ctx, expected, map[string]any{"value": 2}); err == nil {
		t.Fatal("StdCheckError() returned nil for different numeric values")
	}
}
