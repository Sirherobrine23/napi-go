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
	case *_Value:
		if len(v.extra) == 1 {
			promiseDeferred, ok := v.extra[0].(napi.Deferred)
			if ok {
				return &Promise{v, promiseDeferred}
			}
		}
	}
	panic("cannot convert ValueType to Promise, required create by CreatePromise")
}

// CreatePromise creates a new JavaScript Promise object in the given N-API environment.
// It returns a pointer to the created Promise and an error if the operation fails.
// The function internally calls napi.CreatePromise to obtain the promise value and deferred handle.
// If an error occurs during promise creation, it is converted and returned.
func CreatePromise(env EnvType) (*Promise, error) {
	promiseValue, promiseDeferred, err := napi.CreatePromise(env.NapiValue())
	if err := err.ToError(); err != nil {
		return nil, err
	}
	return ToPromise(N_APIValue(env, promiseValue, promiseDeferred)), nil
}

// Reject rejects the promise with the provided value.
// It calls napi.RejectDeferred to reject the underlying N-API deferred promise
// using the given ValueType. Returns an error if the rejection fails.
func (promise *Promise) Reject(value ValueType) error {
	return napi.RejectDeferred(promise.NapiEnv(), promise.promiseDeferred, value.NapiValue()).ToError()
}

// Resolve fulfills the promise with the provided value.
// It resolves the underlying N-API deferred object using the given ValueType.
// Returns an error if the resolution fails.
func (promise *Promise) Resolve(value ValueType) error {
	return napi.ResolveDeferred(promise.NapiEnv(), promise.promiseDeferred, value.NapiValue()).ToError()
}
