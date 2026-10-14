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
			value, err := napi.ToNumber(ci.Args[0]).Int()
			if err != nil {
				return nil, err
			}
			state.value = value
		}
		return nil, napi.ToObject(ci.This).Wrap(state, nil)
	}, napi.PropertyDescriptor{
		Utf8name: "increment", Attributes: napi.DefaultMethod,
		Method: func(ci *napi.CallbackInfo) (napi.ValueType, error) {
			data, err := napi.ToObject(ci.This).Unwrap()
			if err != nil {
				return nil, err
			}
			state := data.(*counter)
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
