// Package module provides the entry point and initialization logic for N-API modules written in Go.
// It handles the integration with Node.js via cgo, sets up the environment, and exposes the
// Register function for module registration.
//
// The initializeModule function is exported for use by Node.js and is responsible for initializing
// the internal N-API environment, converting C types to Go types, and invoking the Register function
// to attach module exports. If a panic occurs during registration, it is caught and an error is
// thrown in the JavaScript environment.
//
// The Register function must be linked using go:linkname and is intended to be implemented by the
// user to define the module's exported functions and properties. See the provided example in the
// comments for usage details.
package module

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

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	internal_napi "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

//export initializeModule
func initializeModule(cEnv C.napi_env, cExports C.napi_value) C.napi_value {
	// Start cgo internal napi values
	internalEnv, internalExports := internal_napi.Env(cEnv), internal_napi.Value(cExports)
	internal_napi.InitializeInstanceData(internalEnv)

	// Convert to go type
	env := napi.N_APIEnv(internalEnv)
	export := napi.ToObject(napi.N_APIValue(env, internalExports))

	// Recover if have panic in Register
	defer func() {
		if err := recover(); err != nil {
			switch v := err.(type) {
			case error:
				internal_napi.ThrowError(env.NapiValue(), "", v.Error())
			default:
				internal_napi.ThrowError(env.NapiValue(), "", fmt.Sprintf("%s", v))
			}
		}
	}()

	// Call register
	Register(env, export)

	// return value
	return cExports
}

// Function to register N-API module functions and other on export Object,
// this function require use go:linkname to link register function.
// Se https://pkg.go.dev/cmd/compile#hdr-Linkname_Directive to how link Register function.
// If have panic in register call, throw error in javascript.
//
// # register example:
//
//	package main
//
//	import _ "unsafe" 																						// Require to go:linkname
//	import _ "sirherobrine23.com.br/Sirherobrine23/napi-go/module" // Module register import
//
//	import "sirherobrine23.com.br/Sirherobrine23/napi-go"
//
//	func main() {}
//
//	//go:linkname register sirherobrine23.com.br/Sirherobrine23/napi-go/entry.Register
//	func register(env napi.EnvType, export *napi.Object) {
//		str, _ := napi.CreateString(env, "hello from Gopher")
//		export.Set("msg", str)
//	}
//
//go:linkname Register
func Register(env napi.EnvType, export *napi.Object)
