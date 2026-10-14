package main

import (
	_ "unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	_ "sirherobrine23.com.br/Sirherobrine23/napi-go/module"
)

type counter struct{ value int64 }

//go:linkname Register sirherobrine23.com.br/Sirherobrine23/napi-go/module.Register
func Register(env napi.EnvType, export *napi.Object) {
	class, err := napi.CreateClass(env, "Counter", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		state := &counter{}
		if len(ci.Args) > 0 {
			if err := ci.ArgInto(0, &state.value); err != nil {
				return nil, err
			}
		}
		return nil, napi.ToObject(ci.This).Wrap(state, nil)
	}, napi.PropertyDescriptor{
		Utf8name: "increment", Attributes: napi.DefaultMethod,
		Method: func(ci *napi.CallbackInfo) (napi.ValueType, error) {
			state, err := napi.UnwrapAs[*counter](napi.ToObject(ci.This))
			if err != nil {
				return nil, err
			}
			state.value++
			return napi.CreateNumber(ci.Env, state.value)
		},
	})
	if err != nil {
		panic(err)
	}
	if err := export.Set("Counter", class); err != nil {
		panic(err)
	}
}

func main() {}
