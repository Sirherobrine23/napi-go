package napi

/*
#include <node/node_api.h>
*/
import "C"
import "unsafe"

type PropertyDescriptor struct {
	Utf8name   string
	Name       Value
	Method     Callback
	Getter     Callback
	Setter     Callback
	Value      Value
	Attributes PropertyAttributes
	Data       unsafe.Pointer
}

type Finalize func(env Env, finalizeData, finalizeHint unsafe.Pointer)

func FinalizeToFinalizer(fn Finalize) Finalizer {
	return func(env C.napi_env, finalizeData, finalizeHint unsafe.Pointer) {
		fn(Env(env), finalizeData, finalizeHint)
	}
}

// Finalizer as a C-compatible function pointer type
type Finalizer func(env C.napi_env, finalizeData, finalizeHint unsafe.Pointer)

type Reference struct {
	Ref unsafe.Pointer
}

type EscapableHandleScope struct {
	Scope unsafe.Pointer
}

type HandleScope struct {
	Scope unsafe.Pointer
}
