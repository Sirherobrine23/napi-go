package main

import (
	_ "unsafe"

	_ "sirherobrine23.com.br/Sirherobrine23/napi-go/entry"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
)

func main() {}

//go:linkname Register sirherobrine23.com.br/Sirherobrine23/napi-go/entry.Register
func Register(env napi.EnvType, export *napi.Object) {
	class, err := napi.CreateClass[*ClassTest](env, []napi.PropertyDescriptor{})
	if err != nil {
		panic(err)
	}
	
	export.Set("class", class)
}

type ClassTest struct{}

func (class *ClassTest) Contructor(env napi.EnvType, this napi.ValueType, args []napi.ValueType) (*napi.Object, error) {
	obj, _ := napi.CreateObject(env)

	return obj, nil
}
