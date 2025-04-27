package napi

// #include <node/node_api.h>
import "C"

import "unsafe"

// Function pointer type for add-on provided function that allow the user to schedule a group of calls to Node-APIs in response to a garbage collection event, after the garbage collection cycle has completed. These function pointers can be used with node_api_post_finalizer.
type Finalize func(env Env, finalizeData, finalizeHint unsafe.Pointer)

// Finalizer as a C-compatible function pointer type
type Finalizer func(env C.napi_env, finalizeData, finalizeHint unsafe.Pointer)

func FinalizeToFinalizer(fn Finalize) Finalizer {
	return func(env C.napi_env, finalizeData, finalizeHint unsafe.Pointer) {
		fn(Env(env), finalizeData, finalizeHint)
	}
}

type Reference struct {
	Ref unsafe.Pointer
}

func CreateReference(env Env, value Value, initialRefcount int) (Reference, Status) {
	var ref Reference
	status := Status(C.napi_create_reference(
		C.napi_env(env),
		C.napi_value(value),
		C.uint32_t(initialRefcount),
		(*C.napi_ref)(unsafe.Pointer(&ref)),
	))
	return ref, status
}

func DeleteReference(env Env, ref Reference) Status {
	return Status(C.napi_delete_reference(
		C.napi_env(env),
		C.napi_ref(unsafe.Pointer(&ref)),
	))
}

func ReferenceRef(env Env, ref Reference) (int, Status) {
	var result C.uint32_t
	status := Status(C.napi_reference_ref(
		C.napi_env(env),
		C.napi_ref(unsafe.Pointer(&ref)),
		&result,
	))
	return int(result), status
}

func ReferenceUnref(env Env, ref Reference) (int, Status) {
	var result C.uint32_t
	status := Status(C.napi_reference_unref(
		C.napi_env(env),
		C.napi_ref(ref.Ref),
		&result,
	))
	return int(result), status
}

func GetReferenceValue(env Env, ref Reference) (Value, Status) {
	var result Value
	status := Status(C.napi_get_reference_value(
		C.napi_env(env),
		C.napi_ref(unsafe.Pointer(&ref)),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

// Wrap function to use the Finalizer type internally
func Wrap(env Env, jsObject Value, nativeObject unsafe.Pointer, finalize Finalize, finalizeHint unsafe.Pointer) Status {
	var result Reference
	status := Status(C.napi_wrap(
		C.napi_env(env),
		C.napi_value(jsObject),
		nativeObject,
		C.napi_finalize(unsafe.Pointer(&finalize)),
		finalizeHint,
		(*C.napi_ref)(unsafe.Pointer(&result)),
	))
	return status
}

func Unwrap(env Env, jsObject Value) (unsafe.Pointer, Status) {
	var nativeObject unsafe.Pointer
	status := Status(C.napi_unwrap(
		C.napi_env(env),
		C.napi_value(jsObject),
		&nativeObject,
	))
	return nativeObject, status
}

func RemoveWrap(env Env, jsObject Value) Status {
	var result unsafe.Pointer
	return Status(C.napi_remove_wrap(
		C.napi_env(env),
		C.napi_value(jsObject),
		&result,
	))
}

type EscapableHandleScope struct {
	Scope unsafe.Pointer
}

type HandleScope struct {
	Scope unsafe.Pointer
}

func OpenHandleScope(env Env) (HandleScope, Status) {
	var scope HandleScope
	status := Status(C.napi_open_handle_scope(
		C.napi_env(env),
		(*C.napi_handle_scope)(unsafe.Pointer(&scope)),
	))
	return scope, status
}

func CloseHandleScope(env Env, scope HandleScope) Status {
	return Status(C.napi_close_handle_scope(
		C.napi_env(env),
		C.napi_handle_scope(unsafe.Pointer(&scope)),
	))
}

func OpenEscapableHandleScope(env Env) (EscapableHandleScope, Status) {
	var scope EscapableHandleScope
	status := Status(C.napi_open_escapable_handle_scope(
		C.napi_env(env),
		(*C.napi_escapable_handle_scope)(unsafe.Pointer(&scope)),
	))
	return scope, status
}

func CloseEscapableHandleScope(env Env, scope EscapableHandleScope) Status {
	return Status(C.napi_close_escapable_handle_scope(
		C.napi_env(env),
		C.napi_escapable_handle_scope(unsafe.Pointer(&scope)),
	))
}

func EscapeHandle(env Env, scope EscapableHandleScope, escapee Value) (Value, Status) {
	var result Value
	status := Status(C.napi_escape_handle(
		C.napi_env(env),
		C.napi_escapable_handle_scope(unsafe.Pointer(&scope)),
		C.napi_value(escapee),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}
