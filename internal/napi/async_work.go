package napi

// #include <stdlib.h>
// #include <node/node_api.h>
import "C"

import "unsafe"

type AsyncWork struct {
	Handle unsafe.Pointer
	ID     NapiGoAsyncWorkID
}

type AsyncExecuteCallback func(env Env)

type AsyncCompleteCallback func(env Env, status Status)

func CreateAsyncWork(env Env, asyncResource, asyncResourceName Value, execute AsyncExecuteCallback, complete AsyncCompleteCallback) (AsyncWork, Status) {
	provider, status := getInstanceData(env)
	if status != StatusOK || provider == nil {
		return AsyncWork{}, status
	}

	return provider.GetAsyncWorkData().CreateAsyncWork(env, asyncResource, asyncResourceName, execute, complete)
}

func DeleteAsyncWork(env Env, work AsyncWork) Status {
	provider, status := getInstanceData(env)
	if status != StatusOK || provider == nil {
		return status
	}

	defer provider.GetAsyncWorkData().DeleteAsyncWork(work.ID)
	return Status(C.napi_delete_async_work(
		C.napi_env(env),
		C.napi_async_work(work.Handle),
	))
}

func QueueAsyncWork(env Env, work AsyncWork) Status {
	return Status(C.napi_queue_async_work(
		C.napi_env(env),
		C.napi_async_work(work.Handle),
	))
}

func CancelAsyncWork(env Env, work AsyncWork) Status {
	return Status(C.napi_cancel_async_work(
		C.napi_env(env),
		C.napi_async_work(work.Handle),
	))
}
