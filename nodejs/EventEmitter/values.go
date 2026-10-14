package event

import (
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs"
)

// EmitValues converts Go arguments and emits synchronously on the Node thread.
// For background Go producers use EventEmitterThread.Emit instead.
func (eve *EventEmitter) EmitValues(name string, args ...any) (bool, error) {
	values := make([]napi.ValueType, len(args))
	for i, arg := range args {
		value, err := napi.ValueOf(eve.env, arg)
		if err != nil {
			return false, err
		}
		values[i] = value
	}
	return eve.Emit(name, values...)
}

// OnFunc registers a Go or JS listener and returns its JS function so it can be
// removed with RemoveListener in this scope (or by retaining it in JavaScript).
func (eve *EventEmitter) OnFunc(name string, callback any) (*napi.Function, error) {
	fn, err := nodejs.CallbackOf(eve.env, callback)
	if err != nil {
		return nil, err
	}
	if err := eve.On(name, fn); err != nil {
		return nil, err
	}
	return fn, nil
}

func (eve *EventEmitter) OnceFunc(name string, callback any) (*napi.Function, error) {
	fn, err := nodejs.CallbackOf(eve.env, callback)
	if err != nil {
		return nil, err
	}
	if err := eve.Once(name, fn); err != nil {
		return nil, err
	}
	return fn, nil
}
