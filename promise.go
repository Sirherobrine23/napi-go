package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

type Promise struct {
	value
	promiseDeferred napi.Deferred
}

// If [ValueType] is [*Promise] return else panic
func ToPromise(o ValueType) *Promise {
	switch v := o.(type) {
	case *Promise:
		return v
	case *Value:
		if len(v.extra) == 1 {
			promiseDeferred, ok := v.extra[0].(napi.Deferred)
			if ok {
				return &Promise{v, promiseDeferred}
			}
		}
	}
	panic("cannot convert ValueType to Promise, required create by CreatePromise")
}

func CreatePromise(env EnvType) (*Promise, error) {
	promiseValue, promiseDeferred, err := napi.CreatePromise(env.NapiValue())
	if err := err.ToError(); err != nil {
		return nil, err
	}
	return ToPromise(N_APIValue(env, promiseValue, promiseDeferred)), nil
}

func (promise *Promise) Reject(value ValueType) error {
	return napi.RejectDeferred(promise.NapiEnv(), promise.promiseDeferred, value.NapiValue()).ToError()
}

func (promise *Promise) Resolve(value ValueType) error {
	return napi.ResolveDeferred(promise.NapiEnv(), promise.promiseDeferred, value.NapiValue()).ToError()
}
