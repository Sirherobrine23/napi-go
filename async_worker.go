package napi

import (
	"fmt"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type AsyncWorker struct {
	value
	asyncWork       napi.AsyncWork
	promiseDeferred napi.Deferred
}

type (
	// function to run code in background without locker Loop event
	CallbackAsyncWorkerExec func(env EnvType)

	// Funtion to run after exec code
	CallbackAsyncWorkerDone func(env EnvType, Resolve, Reject func(value ValueType))
)

// Create async worker to run in backgroud N-API code and return Promise
// 
// On `exec` function dont storage Napi values create on, save in go values and return js values on `done` call.
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

func (async *AsyncWorker) Cancel() error {
	return napi.CancelAsyncWork(async.NapiEnv(), async.asyncWork).ToError()
}
