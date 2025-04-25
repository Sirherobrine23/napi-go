package main

import (
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/entry"
)

func init() {
	entry.Register(func(e *napi.Env, o *napi.Object) {
		value, _ := napi.CreateString(e, "test")
		o.Set("test", value.Value)

		fn, _ := napi.CreateFunction(e, "testFunc", func(env *napi.Env, this *napi.Value, args []*napi.Value) (*napi.Value, error) {
			if this.IsObject() {
				for keyName := range this.ToObject().Seq() {
					println(keyName)
				}
			}

			return napi.MustCreateString(env, "test").Value, nil
		})

		o.Set("testFunc", fn.Value)
	})
}

func main() {}
