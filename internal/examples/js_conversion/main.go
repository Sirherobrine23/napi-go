package main

import (
	"encoding/json"
	"net/netip"
	_ "unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	_ "sirherobrine23.com.br/Sirherobrine23/napi-go/module"
)

type Test struct {
	Int    int
	String string
	Sub    []any
}

//go:linkname RegisterNapi sirherobrine23.com.br/Sirherobrine23/napi-go/module.Register
func RegisterNapi(env napi.EnvType, export *napi.Object) {
	inNode, _ := napi.CreateString(env, "from golang napi string")
	inNode2, _ := napi.CopyBuffer(env, []byte{1, 0, 244, 21})
	toGoReflect := &Test{
		Int:    14,
		String: "From golang",
		Sub: []any{
			1,
			[]string{"test", "gopher"},
			[]bool{false, true},
			[]int{23, 244, 10, 2024, 2025, 2000},
			map[string]string{"exampleMap": "test"},
			map[int]string{1: "one"},
			map[bool]string{false: "false", true: "true"},
			map[[2]string]string{{"go"}: "example"},
			netip.IPv4Unspecified(),
			netip.IPv6Unspecified(),
			netip.AddrPortFrom(netip.IPv6Unspecified(), 19132),
			nil,
			true,
			false,
			inNode,
			inNode2,
			func() {
				println("called in go")
			},
		},
	}

	napiStruct, err := napi.ValueOf(env, toGoReflect)
	if err != nil {
		panic(err)
	}
	export.Set("goStruct", napiStruct)

	fnCall, err := napi.GoFuncOf(env, func(call ...any) (string, error) {
		d, err := json.MarshalIndent(call, "", "  ")
		if err == nil {
			println(string(d))
		}
		return string(d), err
	})
	if err != nil {
		panic(err)
	}
	export.Set("printAny", fnCall)

	fnCallStruct, err := napi.GoFuncOf(env, func(call ...Test) (string, error) {
		d, err := json.MarshalIndent(call, "", "  ")
		if err == nil {
			println(string(d))
		}
		return string(d), err
	})
	if err != nil {
		panic(err)
	}
	export.Set("printTestStruct", fnCallStruct)
}

func main() {}
