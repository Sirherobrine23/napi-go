package napi

import (
	"fmt"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

// function to run code in background without locker Loop event
type CallbackAsyncWorkerExec func(env EnvType)

// Funtion to run after exec code
type CallbackAsyncWorkerDone func(env EnvType, Resolve, Reject func(value ValueType))

// AsyncWorker encapsulates an asynchronous worker for executing tasks in a separate thread
// using Node.js N-API. It embeds a value type and manages the lifecycle of an async work
// operation, including its associated promise for JavaScript interoperability.
type AsyncWorker struct {
	value
	asyncWork       napi.AsyncWork
	promiseDeferred napi.Deferred
}

// CreateAsyncWorker creates an asynchronous worker that executes the provided exec function in a separate thread,
// and calls the done callback upon completion. It returns an AsyncWorker instance and a promise that can be used
// to track the asynchronous operation from JavaScript.
//
// The promise is resolved or rejected based on the outcome of the async operation. If an error or panic occurs
// during execution, the promise is rejected with the corresponding error. If the async worker is cancelled, the
// promise is also rejected.
func CreateAsyncWorker(env EnvType, exec CallbackAsyncWorkerExec, done CallbackAsyncWorkerDone) (*AsyncWorker, error) {
	promiseResult, err := CreatePromise(env)
	if err != nil {
		return nil, err
	}
	asyncName, _ := CreateString(env, "napi-go/promiseAsyncWorker")
	status, asyncWork := napi.Status(0), napi.AsyncWork{}
	asyncWork, status = napi.CreateAsyncWork(env.NapiValue(), nil, asyncName.NapiValue(),
		func(env napi.Env) {
			defer func() {
				if err2 := recover(); err2 != nil {
					switch v := err2.(type) {
					case error:
						err = v
					default:
						err = fmt.Errorf("recover panic: %s", v)
					}
					return
				}
				ext, status := napi.GetExtendedErrorInfo(env)
				if status.ToError() == nil {
					println(ext.Message)
				}
			}()
			exec(N_APIEnv(env))
		},
		func(env napi.Env, status napi.Status) {
			defer napi.DeleteAsyncWork(env, asyncWork)
			if status == napi.StatusCancelled {
				err, _ := CreateError(N_APIEnv(env), "async worker canceled")
				napi.RejectDeferred(env, promiseResult.promiseDeferred, err.NapiValue())
				return
			} else if err != nil {
				err, _ := CreateError(N_APIEnv(env), err.Error())
				napi.RejectDeferred(env, promiseResult.promiseDeferred, err.NapiValue())
				return
			}
			defer func() {
				if err := recover(); err != nil {
					switch v := err.(type) {
					case error:
						err, _ := CreateError(N_APIEnv(env), v.Error())
						napi.RejectDeferred(env, promiseResult.promiseDeferred, err.NapiValue())
					default:
						err, _ := CreateError(N_APIEnv(env), fmt.Sprintf("recover panic: %s", v))
						napi.RejectDeferred(env, promiseResult.promiseDeferred, err.NapiValue())
					}
				}
			}()
			var calledEnd bool
			defer func() {
				if calledEnd {
					return
				}
				err, _ := CreateError(N_APIEnv(env), "function end and not call resolved")
				napi.RejectDeferred(env, promiseResult.promiseDeferred, err.NapiValue())
			}()
			done(
				N_APIEnv(env),
				func(value ValueType) {
					calledEnd = true
					if value == nil {
						if value, err = N_APIEnv(env).Undefined(); err != nil {
							panic(err)
						}
					}
					napi.ResolveDeferred(env, promiseResult.promiseDeferred, value.NapiValue())
				},
				func(value ValueType) {
					calledEnd = true
					if value == nil {
						if value, err = N_APIEnv(env).Undefined(); err != nil {
							panic(err)
						}
					}
					napi.RejectDeferred(env, promiseResult.promiseDeferred, value.NapiValue())
				},
			)
		})

	// Check error and start worker
	if err := status.ToError(); err != nil {
		return nil, err
	} else if err = napi.QueueAsyncWork(env.NapiValue(), asyncWork).ToError(); err != nil {
		return nil, err
	}

	return &AsyncWorker{
		promiseResult.value,
		asyncWork,
		promiseResult.promiseDeferred,
	}, nil
}

// Cancel attempts to cancel the asynchronous work associated with the AsyncWorker.
// It returns an error if the cancellation fails or if the async work cannot be cancelled.
func (async *AsyncWorker) Cancel() error {
	return napi.CancelAsyncWork(async.NapiEnv(), async.asyncWork).ToError()
}
