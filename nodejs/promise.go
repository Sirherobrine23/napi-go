package nodejs

import (
	"fmt"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
)

// Promise wraps a JavaScript-owned Promise. It has no napi_deferred, does not
// expose Resolve/Reject, and must never be synchronously awaited on the Node
// thread. Return it to JS or attach callbacks; callback handles are fresh scopes.
type Promise struct{ *napi.Object }

func ToPromise(value napi.ValueType) (*Promise, error) {
	if err := checkValue(value); err != nil {
		return nil, err
	}
	kind, err := value.Type()
	if err != nil {
		return nil, err
	}
	if kind != napi.TypePromise {
		return nil, fmt.Errorf("expected Promise, got %s", kind)
	}
	return &Promise{napi.ToObject(value)}, nil
}

func (module *Module) CallPromise(name string, args ...any) (*Promise, error) {
	value, err := module.Call(name, args...)
	if err != nil {
		return nil, err
	}
	return ToPromise(value)
}

// Then attaches Go or JS functions. A literal nil handler is omitted. A Callback
// may return a Go-created napi Promise or a nodejs Promise to adopt its result.
func (promise *Promise) Then(fulfilled, rejected any) (*Promise, error) {
	if fulfilled == nil && rejected == nil {
		return promise, fmt.Errorf("add one of functions")
	}

	args := make([]any, 2)
	for index, handler := range []any{fulfilled, rejected} {
		if handler != nil {
			fn, err := CallbackOf(promise.Env(), handler)
			if err != nil {
				return nil, err
			}
			args[index] = fn
		}
	}
	value, err := promise.CallMethod("then", args...)
	if err != nil {
		return nil, err
	}
	return ToPromise(value)
}

func (promise *Promise) Catch(rejected any) (*Promise, error) {
	if hasCatch, err := promise.Has("catch"); err == nil && hasCatch && rejected != nil {
		fn, err := CallbackOf(promise.Env(), rejected)
		if err != nil {
			return nil, err
		}
		value, err := promise.CallMethod("catch", fn)
		if err != nil {
			return nil, err
		}
		return ToPromise(value)
	}
	return promise.Then(nil, rejected)
}

func (promise *Promise) Finally(callback any) (*Promise, error) {
	fn, err := CallbackOf(promise.Env(), callback)
	if err != nil {
		return nil, err
	}
	value, err := promise.CallMethod("finally", fn)
	if err != nil {
		return nil, err
	}
	return ToPromise(value)
}
