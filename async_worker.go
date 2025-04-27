package napi

import (
	"context"
	"fmt"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type AsyncWorker struct {
	env       EnvType
	asyncWork napi.AsyncWork
}

type CallbackAsyncWorker func(env EnvType, ctx context.Context)

// Create async worker to run in backgroud N-API code
func CreateAyncWorker(env EnvType, ctx context.Context, fn CallbackAsyncWorker) (*AsyncWorker, error) {
	var status napi.Status
	var asyncWork napi.AsyncWork
	ctx, cancel := context.WithCancelCause(ctx)

	// On start exec
	execute := napi.AsyncExecuteCallback(func(env napi.Env) {
		defer func() {
			if err := recover(); err != nil {
				switch v := err.(type) {
				case error:
					ThrowError(N_APIEnv(env), "", v.Error())
				default:
					ThrowError(N_APIEnv(env), "", fmt.Sprintf("recover panic: %s", v))
				}
			}
		}()
		fn(N_APIEnv(env), ctx)
	})

	// End exec
	complete := napi.AsyncCompleteCallback(func(env napi.Env, status napi.Status) {
		defer napi.DeleteAsyncWork(env, asyncWork)
		switch status {
		case napi.StatusOK, napi.StatusCancelled:
			cancel(nil)
		default:
			cancel(status.ToError())
		}
	})

	name, _ := CreateString(env, "go-napi/async_worker")
	asyncWork, status = napi.CreateAsyncWork(env.NapiValue(), nil, name.NapiValue(), execute, complete)
	if err := status.ToError(); err != nil {
		return nil, err
	} else if err = napi.QueueAsyncWork(env.NapiValue(), asyncWork).ToError(); err != nil {
		return nil, err
	}

	return &AsyncWorker{env, asyncWork}, nil
}

func (async *AsyncWorker) Cancel() error {
	return napi.CancelAsyncWork(async.env.NapiValue(), async.asyncWork).ToError()
}
