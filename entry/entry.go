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
#cgo (darwin && amd64) LDFLAGS: -arch x86_64
#cgo (darwin && arm64) LDFLAGS: -arch arm64

#cgo linux LDFLAGS: -Wl,-unresolved-symbols=ignore-all

#cgo LDFLAGS: -L${SRCDIR}

#include <stdlib.h>
#include "./entry.h"
*/
import "C"

import (
	"fmt"
	_ "unsafe"

	gonapi "sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

//export initializeModule
func initializeModule(cEnv C.napi_env, cExports C.napi_value) C.napi_value {
	env, exports := napi.Env(cEnv), napi.Value(cExports)
	napi.InitializeInstanceData(env)
	newEnv := gonapi.N_APIEnv(env)
	exportObj := gonapi.ToObject(gonapi.N_APIValue(newEnv, exports))

	defer func() {
		if err := recover(); err != nil {
			switch v := err.(type) {
			case error:
				gonapi.ThrowError(gonapi.N_APIEnv(env), "", v.Error())
			default:
				gonapi.ThrowError(gonapi.N_APIEnv(env), "", fmt.Sprintf("%s", v))
			}
		}
		Register(newEnv, exportObj)
	}()
	return cExports
}

// N-API Module register
//
// Se https://pkg.go.dev/cmd/compile#hdr-Linkname_Directive to how link Register function
//
// example:
//
//	package main
//
//	import _ "unsafe"
//	import _ "sirherobrine23.com.br/Sirherobrine23/napi-go/entry"
//	import "sirherobrine23.com.br/Sirherobrine23/napi-go"
//
//	//go:linkname wg sirherobrine23.com.br/Sirherobrine23/napi-go/entry.Register
//	func wg(env napi.EnvType, export *napi.Object) {
//		str, _ := napi.CreateString(env, "hello from Gopher")
//		export.Set("msg", str)
//	}
//
//go:linkname Register
func Register(env gonapi.EnvType, export *gonapi.Object)
