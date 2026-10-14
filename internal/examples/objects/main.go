package main

import (
	"fmt"
	_ "unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	_ "sirherobrine23.com.br/Sirherobrine23/napi-go/module"
)

type config struct {
	Name    string `napi:"name"`
	Enabled bool   `napi:"enabled"`
	Items   []int  `napi:"items"`
}

func createCounter(ci *napi.CallbackInfo) (napi.ValueType, error) {
	start := 0
	if len(ci.Args) > 0 {
		if err := ci.ArgInto(0, &start); err != nil {
			return nil, err
		}
	}
	object, err := napi.ObjectOf(ci.Env, map[string]any{"count": start, "name": "counter"})
	if err != nil {
		return nil, err
	}
	if err := object.SetMethod("increment", func(call *napi.CallbackInfo) (napi.ValueType, error) {
		amount := 1
		if len(call.Args) > 0 {
			if err := call.ArgInto(0, &amount); err != nil {
				return nil, err
			}
		}
		// Retrieve this in every callback; do not capture the creation scope's handle.
		this := napi.ToObject(call.This)
		count, err := napi.GetAs[int](this, "count")
		if err != nil {
			return nil, err
		}
		if err := this.SetValue("count", count+amount); err != nil {
			return nil, err
		}
		return napi.ValueOf(call.Env, count+amount)
	}); err != nil {
		return nil, err
	}
	return object, nil
}

//go:linkname Register sirherobrine23.com.br/Sirherobrine23/napi-go/module.Register
func Register(env napi.EnvType, export *napi.Object) {
	if err := export.Assign(map[string]any{"version": "1", "defaults": config{Name: "Gopher", Enabled: true, Items: []int{1, 2}}}); err != nil {
		panic(err)
	}
	if err := export.SetMethod("createCounter", createCounter); err != nil {
		panic(err)
	}
	if err := export.SetMethod("describe", func(input config) string {
		return fmt.Sprintf("%s: enabled=%t, items=%d", input.Name, input.Enabled, len(input.Items))
	}); err != nil {
		panic(err)
	}
}

func main() {}
