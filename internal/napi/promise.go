package napi

// #include <stdlib.h>
// #include <node/node_api.h>
import "C"

import "unsafe"

type Deferred unsafe.Pointer

type Promise struct {
	Deferred Deferred
	Value    Value
}

// This API creates a deferred object and a JavaScript promise.
func CreatePromise(env Env) (Promise, Status) {
	var result Promise
	status := Status(C.napi_create_promise(
		C.napi_env(env),
		(*C.napi_deferred)(unsafe.Pointer(&result.Deferred)),
		(*C.napi_value)(unsafe.Pointer(&result.Value)),
	))
	return result, status
}

// This API resolves a JavaScript promise by way of the deferred object with which it is associated.
// Thus, it can only be used to resolve JavaScript promises for which the corresponding deferred object is available.
// This effectively means that the promise must have been created using napi_create_promise() and the deferred object returned from that call must have been retained in order to be passed to this API.
func ResolveDeferred(env Env, deferred Deferred, resolution Value) Status {
	return Status(C.napi_resolve_deferred(
		C.napi_env(env),
		C.napi_deferred(deferred),
		C.napi_value(resolution),
	))
}

// This API rejects a JavaScript promise by way of the deferred object with which it is associated.
// Thus, it can only be used to reject JavaScript promises for which the corresponding deferred object is available.
// This effectively means that the promise must have been created using napi_create_promise() and the deferred object returned from that call must have been retained in order to be passed to this API.
func RejectDeferred(env Env, deferred Deferred, rejection Value) Status {
	return Status(C.napi_reject_deferred(
		C.napi_env(env),
		C.napi_deferred(deferred),
		C.napi_value(rejection),
	))
}
