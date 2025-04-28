package main

import (
	"time"
	_ "unsafe"

	_ "sirherobrine23.com.br/Sirherobrine23/napi-go/entry"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
)

var waitTime = time.Second * 3

//go:linkname Register sirherobrine23.com.br/Sirherobrine23/napi-go/entry.Register
func Register(env napi.EnvType, export *napi.Object) {
	thr, err := napi.CreateThreadsafeFunction(
		env,
		func(env napi.EnvType, this napi.ValueType, args []napi.ValueType) (napi.ValueType, error) {
			return nil, nil
		},
		"thr",
		0,
		1,
		func(env napi.EnvType, jsCallback *napi.Function, data any) {
			println("called 2")
		},
		nil,
		nil,
	)
	if err != nil {
		panic(err)
	}
	export.Set("thr", thr)
}

func main() {}
