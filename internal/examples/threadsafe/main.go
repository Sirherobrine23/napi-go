package main

import (
	"fmt"
	"time"
	_ "unsafe"

	_ "sirherobrine23.com.br/Sirherobrine23/napi-go/module"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
)

//go:linkname Register sirherobrine23.com.br/Sirherobrine23/napi-go/module.Register
func Register(env napi.EnvType, export *napi.Object) {
	jsFunc := napi.Callback(func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		var waitTime = time.Second * 3

		if len(ci.Args) == 1 {
			if typof, _ := ci.Args[0].Type(); typof == napi.TypeNumber {
				wait := napi.As[*napi.Number](ci.Args[0])
				_waitTime, _ := wait.Int()
				waitTime = time.Duration(_waitTime)
			}
		}

		fmt.Printf("Called JS, waiting %s\n", waitTime)
		<-time.After(waitTime)
		return nil, nil
	})

	jsEnd := napi.ThreadsafeFunctionFinalizeCallback(func(env napi.EnvType, context any) {
		println("Called go func end")
	})

	callJSCallback := napi.ThreadsafeFunctionCallJSCallback(func(env napi.EnvType, jsCallback *napi.Function, data any) {
		println("Called callJSCallback")
	})

	thr, err := napi.CreateThreadsafeFunction(env, jsFunc, jsEnd, callJSCallback, "thr", 0, 1, nil)
	if err != nil {
		panic(err)
	}
	export.Set("thr", thr)
}

func main() {}
