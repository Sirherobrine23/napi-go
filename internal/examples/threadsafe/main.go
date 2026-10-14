package main

import (
	"fmt"
	_ "unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	_ "sirherobrine23.com.br/Sirherobrine23/napi-go/module"
)

//go:linkname Register sirherobrine23.com.br/Sirherobrine23/napi-go/module.Register
func Register(env napi.EnvType, export *napi.Object) {
	start, err := napi.CreateFunction(env, "start", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		if len(ci.Args) != 1 {
			return nil, fmt.Errorf("start requires a JavaScript callback")
		}
		fn, err := napi.CreateThreadsafeFunction(ci.Env, ci.Args[0], nil, nil, "napi-go/example", 4, 1, nil)
		if err != nil {
			return nil, err
		}
		go func() {
			defer fn.Release(napi.Release)
			for i := 0; i < 10; i++ {
				if err := fn.Call(i, napi.Blocking); err != nil {
					return
				}
			}
		}()
		return nil, nil
	})
	if err != nil {
		panic(err)
	}
	if err := export.Set("start", start); err != nil {
		panic(err)
	}
}

func main() {}
