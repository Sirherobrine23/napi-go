package main

// Require to go:linkname and module register import
import (
	_ "unsafe"

	_ "sirherobrine23.com.br/Sirherobrine23/napi-go/module"
	event "sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/EventEmitter"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
)

func main() {}

//go:linkname register sirherobrine23.com.br/Sirherobrine23/napi-go/module.Register
func register(env napi.EnvType, export *napi.Object) {
	en, err := event.CreateEvent(env)
	if err != nil {
		panic(err)
	}

	export.Set("events", en)
}
