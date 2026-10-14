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
// Both handlers may be nil: the new chain forwards the original settlement.
func (promise *Promise) Then(fulfilled, rejected any) (*Promise, error) {
	return promise.chain("then", fulfilled, rejected)
}

// Catch invokes this Promise's catch method. A literal nil handler forwards
// the original settlement to the returned chain.
func (promise *Promise) Catch(rejected any) (*Promise, error) {
	return promise.chain("catch", rejected)
}

// Finally invokes this Promise's finally method. A literal nil callback forwards
// the original settlement without running cleanup.
func (promise *Promise) Finally(callback any) (*Promise, error) {
	return promise.chain("finally", callback)
}

func (promise *Promise) chain(method string, handlers ...any) (*Promise, error) {
	args := make([]any, len(handlers))
	for index, handler := range handlers {
		if handler != nil {
			fn, err := CallbackOf(promise.Env(), handler)
			if err != nil {
				return nil, err
			}
			args[index] = fn
		}
	}
	// Call directly: checking Has first can invoke extra JS traps, and falling
	// back to another method would hide missing methods or lookup failures.
	value, err := promise.CallMethod(method, args...)
	if err != nil {
		return nil, err
	}
	return ToPromise(value)
}
