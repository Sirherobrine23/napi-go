package napi

// #include <node/node_api.h>
import "C"

import "unsafe"

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
