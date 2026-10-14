package napi

// #include "managed_callbacks.h"
import "C"

import (
	"fmt"
	"runtime/cgo"
	"unsafe"
)

type (
	ThreadsafeFunction            unsafe.Pointer
	ThreadsafeFunctionReleaseMode C.napi_threadsafe_function_release_mode
	ThreadsafeFunctionCallMode    C.napi_threadsafe_function_call_mode
)

const (
	Release     ThreadsafeFunctionReleaseMode = C.napi_tsfn_release
	Abort       ThreadsafeFunctionReleaseMode = C.napi_tsfn_abort
	NonBlocking ThreadsafeFunctionCallMode    = C.napi_tsfn_nonblocking
	Blocking    ThreadsafeFunctionCallMode    = C.napi_tsfn_blocking
)

func CreateThreadsafeFunction(env Env, fn, asyncResource, asyncResourceName Value, maxQueueSize, initialThreadCount int) (ThreadsafeFunction, Status) {
	var result ThreadsafeFunction
	status := Status(C.napi_create_threadsafe_function(
		C.napi_env(env),
		C.napi_value(fn),
		C.napi_value(asyncResource),
		C.napi_value(asyncResourceName),
		C.size_t(maxQueueSize),
		C.size_t(initialThreadCount),
		nil,
		nil,
		nil,
		nil,
		(*C.napi_threadsafe_function)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CallThreadsafeFunction(fn ThreadsafeFunction, mode ThreadsafeFunctionCallMode) Status {
	return Status(C.napi_call_threadsafe_function(
		C.napi_threadsafe_function(fn),
		nil,
		C.napi_threadsafe_function_call_mode(mode),
	))
}

type ThreadsafeCallJS func(env Env, callback Value, data any)
type ThreadsafeFinalize func(env Env)

type managedThreadsafeCallbacks struct {
	call     ThreadsafeCallJS
	finalize ThreadsafeFinalize
}

func CreateThreadsafeFunctionWithCallbacks(env Env, fn, name Value, maxQueueSize, initialThreadCount int, call ThreadsafeCallJS, finalize ThreadsafeFinalize) (ThreadsafeFunction, Status) {
	token := newToken(managedThreadsafeCallbacks{call, finalize})
	if token == nil {
		return nil, StatusGenericFailure
	}
	var result C.napi_threadsafe_function
	status := Status(C.napi_create_threadsafe_function(C.napi_env(env), C.napi_value(fn), nil,
		C.napi_value(name), C.size_t(maxQueueSize), C.size_t(initialThreadCount), unsafe.Pointer(token),
		C.napi_finalize(C.napi_go_tsfn_finalize), unsafe.Pointer(token),
		C.napi_threadsafe_function_call_js(C.napi_go_tsfn_call), &result))
	if status != StatusOK {
		deleteToken(token)
	}
	return ThreadsafeFunction(result), status
}

//export executeManagedThreadsafeCall
func executeManagedThreadsafeCall(cEnv C.napi_env, fn C.napi_value, context, data C.uintptr_t) {
	var value any
	if data != 0 {
		handle := cgo.Handle(data)
		value = handle.Value().(struct{ Data any }).Data
		handle.Delete()
	}
	// Node can drain queued items without an environment during teardown.
	// Delete their Go handles, but do not run JavaScript or user callbacks.
	if cEnv == nil {
		return
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			ThrowError(Env(cEnv), "", fmt.Sprintf("threadsafe callback panic: %v", recovered))
		}
	}()
	callbacks := cgo.Handle(context).Value().(managedThreadsafeCallbacks)
	callbacks.call(Env(cEnv), Value(fn), value)
}

//export finalizeManagedThreadsafeFunction
func finalizeManagedThreadsafeFunction(cEnv C.napi_env, context C.uintptr_t) {
	handle := cgo.Handle(context)
	defer handle.Delete()
	defer func() {
		if recovered := recover(); recovered != nil && cEnv != nil {
			ThrowError(Env(cEnv), "", fmt.Sprintf("threadsafe finalizer panic: %v", recovered))
		}
	}()
	callbacks := handle.Value().(managedThreadsafeCallbacks)
	callbacks.finalize(Env(cEnv))
}

func CallThreadsafeFunctionWithData(fn ThreadsafeFunction, data any, mode ThreadsafeFunctionCallMode) Status {
	// Store nil inside a struct: a cgo.Handle must contain a non-nil entry.
	token := newToken(struct{ Data any }{data})
	if token == nil {
		return StatusGenericFailure
	}
	status := Status(C.napi_call_threadsafe_function(C.napi_threadsafe_function(fn), unsafe.Pointer(token), C.napi_threadsafe_function_call_mode(mode)))
	if status != StatusOK {
		deleteToken(token)
	}
	return status
}

func AcquireThreadsafeFunction(fn ThreadsafeFunction) Status {
	return Status(C.napi_acquire_threadsafe_function(
		C.napi_threadsafe_function(fn),
	))
}

func ReleaseThreadsafeFunction(fn ThreadsafeFunction, mode ThreadsafeFunctionReleaseMode) Status {
	return Status(C.napi_release_threadsafe_function(
		C.napi_threadsafe_function(fn),
		C.napi_threadsafe_function_release_mode(mode),
	))
}

func GetThreadsafeFunctionContext(fn ThreadsafeFunction) (unsafe.Pointer, Status) {
	var context unsafe.Pointer
	status := Status(C.napi_get_threadsafe_function_context(
		C.napi_threadsafe_function(fn),
		&context,
	))
	return context, status
}

func RefThreadsafeFunction(env Env, fn ThreadsafeFunction) Status {
	return Status(C.napi_ref_threadsafe_function(
		C.napi_env(env),
		C.napi_threadsafe_function(fn),
	))
}

func UnrefThreadsafeFunction(env Env, fn ThreadsafeFunction) Status {
	return Status(C.napi_unref_threadsafe_function(
		C.napi_env(env),
		C.napi_threadsafe_function(fn),
	))
}
