package main

import (
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	entry "sirherobrine23.com.br/Sirherobrine23/napi-go/module/binding"
)

func init() {
	entry.Register(func(env napi.EnvType, export *napi.Object) {
		str, _ := napi.CreateString(env, "from golang")
		export.Set("from", str)
	})
}

func main() {}
