// N-API used by [sirherobrine23.com.br/Sirherobrine23/napi-go], only use in internal, don't export
package napi

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
*/
import "C"

import "unsafe"

type (
	// napi_env is used to represent a context that the underlying Node-API implementation can use to persist VM-specific state. This structure is passed to native functions when they're invoked, and it must be passed back when making Node-API calls. Specifically, the same napi_env that was passed in when the initial native function was called must be passed to any subsequent nested Node-API calls. Caching the napi_env for the purpose of general reuse, and passing the napi_env between instances of the same addon running on different Worker threads is not allowed. The napi_env becomes invalid when an instance of a native addon is unloaded. Notification of this event is delivered through the callbacks given to napi_add_env_cleanup_hook and napi_set_instance_data.
	Env unsafe.Pointer

	// This is an opaque pointer that is used to represent a JavaScript value.
	Value unsafe.Pointer
)
