package main

import (
	"context"
	"encoding/json"
	_ "unsafe"

	_ "sirherobrine23.com.br/Sirherobrine23/napi-go/entry"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/js"
)

//go:linkname Register sirherobrine23.com.br/Sirherobrine23/napi-go/entry.Register
func Register(env napi.EnvType, export *napi.Object) {
	fn, _ := napi.CreateFunction(env, "", func(env napi.EnvType, _ napi.ValueType, args []napi.ValueType) (napi.ValueType, error) {
		var a []any
		for _, arg := range args {
			var b any
			if err := js.ValueFrom(arg, &b); err != nil {
				return nil, err
			}
			a = append(a, b)
		}
		ctx := context.WithValue(context.Background(), "this", a)
		async, err := napi.CreateAyncWorker(env, ctx, func(env napi.EnvType, ctx context.Context) {
			a := ctx.Value("this")
			d, _ := json.MarshalIndent(a, "", "  ")
			println(string(d))
		})
		_ = async

		return nil, err
	})
	export.Set("async", fn)
}

func main() {}
