package main

import (
	_ "unsafe"

	_ "sirherobrine23.com.br/Sirherobrine23/napi-go/module"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
)

func main() {}

//go:linkname Register sirherobrine23.com.br/Sirherobrine23/napi-go/module.Register
func Register(env napi.EnvType, export *napi.Object) {
	class, err := napi.CreateClass[*ClassTest](env)
	if err != nil {
		panic(err)
	}

	export.Set("class", class)
}

type ClassTest struct{}

func (class *ClassTest) Contructor(ci *napi.CallbackInfo) (*napi.Object, error) {
	obj, _ := napi.CreateObject(ci.Env)

	return obj, nil
}
