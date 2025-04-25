package entry

/*
#cgo CFLAGS: -DDEBUG
#cgo CFLAGS: -D_DEBUG
#cgo CFLAGS: -DV8_ENABLE_CHECKS
#cgo CFLAGS: -DNAPI_EXPERIMENTAL
#cgo CFLAGS: -I/usr/local/include/node
#cgo CXXFLAGS: -std=c++11

#cgo darwin LDFLAGS: -Wl,-undefined,dynamic_lookup
#cgo darwin LDFLAGS: -Wl,-no_pie
#cgo darwin LDFLAGS: -Wl,-search_paths_first
// #cgo darwin amd64 LDFLAGS: -arch x86_64
// #cgo darwin arm64 LDFLAGS: -arch arm64

#cgo linux LDFLAGS: -Wl,-unresolved-symbols=ignore-all

#cgo LDFLAGS: -L${SRCDIR}

#include <stdlib.h>
#include "./entry.h"
*/
import "C"

import (
	gonapi "sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type registerCallback func(env napi.Env, object napi.Value)

var modFuncInit = []registerCallback{}

//export initializeModule
func initializeModule(cEnv C.napi_env, cExports C.napi_value) C.napi_value {
	env, exports := napi.Env(cEnv), napi.Value(cExports)
	napi.InitializeInstanceData(env)

	for _, registerCall := range modFuncInit {
		registerCall(env, exports)
	}

	// for name, functionCall := range modFuncInit {
	// 	cb, _ := napi.CreateFunction(env, name, functionCall)
	// 	name, _ := napi.CreateStringUtf8(env, name)
	// 	napi.SetProperty(env, exports, name, cb)
	// }
	return cExports
}

func Register(fn func(*gonapi.Env, *gonapi.Object)) {
	modFuncInit = append(modFuncInit, func(env napi.Env, object napi.Value) {
		registerEnv := gonapi.FromEnvNapi(env)
		registerObj := gonapi.FromValueNapi(registerEnv, object).ToObject()
		fn(registerEnv, registerObj)
	})
}
