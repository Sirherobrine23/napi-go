package napi

/*
#include <node/node_api.h>
*/
import "C"

type ThreadsafeFunctionReleaseMode C.napi_threadsafe_function_release_mode

const (
	Release ThreadsafeFunctionReleaseMode = C.napi_tsfn_release
	Abort   ThreadsafeFunctionReleaseMode = C.napi_tsfn_abort
)

type ThreadsafeFunctionCallMode C.napi_threadsafe_function_call_mode

const (
	NonBlocking ThreadsafeFunctionCallMode = C.napi_tsfn_nonblocking
	Blocking    ThreadsafeFunctionCallMode = C.napi_tsfn_blocking
)
