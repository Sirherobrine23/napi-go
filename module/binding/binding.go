// Export values to Javascript
//
// Deprecated: Use [sirherobrine23.com.br/Sirherobrine23/napi-go/entry] to linking, in final release this module ar to remove
package entry

import (
	_ "unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	_ "sirherobrine23.com.br/Sirherobrine23/napi-go/module"
)

type registerCallback func(env napi.EnvType, object *napi.Object)

var modFuncInit = []registerCallback{}

//go:linkname start sirherobrine23.com.br/Sirherobrine23/napi-go/module.Register
func start(env napi.EnvType, export *napi.Object) {
	for _, registerCall := range modFuncInit {
		registerCall(env, export)
	}
}

// Register callback to register export values
func Register(fn registerCallback) {
	modFuncInit = append(modFuncInit, fn)
}
