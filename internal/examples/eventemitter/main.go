package main

import (
	"fmt"
	_ "unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	_ "sirherobrine23.com.br/Sirherobrine23/napi-go/module"
)

type event struct {
	Name string `napi:"name"`
	Data any    `napi:"data"`
}

func start(ci *napi.CallbackInfo) (napi.ValueType, error) {
	if len(ci.Args) != 2 {
		return nil, fmt.Errorf("start requires a callback and an item count")
	}
	count, err := napi.ToNumber(ci.Args[1]).Int()
	if err != nil {
		return nil, err
	}
	if count < 0 || count > 10000 {
		return nil, fmt.Errorf("item count must be between 0 and 10000")
	}
	emit, err := napi.CreateThreadsafeFunction(ci.Env, ci.Args[0], nil, nil, "eventemitter/example", 64, 1, nil)
	if err != nil {
		return nil, err
	}
	go func() {
		defer emit.Release(napi.Release)
		defer func() {
			if recovered := recover(); recovered != nil {
				_ = emit.Call(event{"error", fmt.Sprint(recovered)}, napi.Blocking)
			}
		}()
		var sum int64
		for index := 0; index < int(count); index++ {
			value := index * index
			sum += int64(value)
			if err := emit.Call(event{"progress", map[string]int{"index": index, "value": value}}, napi.Blocking); err != nil {
				return
			}
		}
		_ = emit.Call(event{"done", map[string]any{"count": int(count), "sum": float64(sum)}}, napi.Blocking)
	}()
	return nil, nil
}

//go:linkname Register sirherobrine23.com.br/Sirherobrine23/napi-go/module.Register
func Register(env napi.EnvType, export *napi.Object) {
	fn, err := napi.CreateFunction(env, "start", start)
	if err != nil {
		panic(err)
	}
	if err := export.Set("start", fn); err != nil {
		panic(err)
	}
}

func main() {}
