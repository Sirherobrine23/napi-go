package event

import (
	"fmt"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/require"

	internalnapi "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

var _ napi.ValueType = (*EventEmitter)(nil)

func CreateEvent(env napi.EnvType) (*EventEmitter, error) {
	events, err := require.Builtin(env, "node:events")
	if err != nil {
		return nil, fmt.Errorf("cannot get event module: %w", err)
	}
	jsEventEmitter, err := napi.ToObject(events).Get("EventEmitter")
	if err != nil {
		return nil, err
	}
	newClass, err := napi.ToClass(jsEventEmitter).New()
	if err != nil {
		return nil, err
	}

	return &EventEmitter{
		env:     env,
		valueOf: napi.ToObject(newClass),
	}, nil
}

type EventEmitter struct {
	env     napi.EnvType // Env
	valueOf *napi.Object // JS value
}

func (eve *EventEmitter) Env() napi.EnvType             { return eve.env }
func (eve *EventEmitter) Type() (napi.NapiType, error)  { return napi.TypeObject, nil }
func (eve *EventEmitter) NapiValue() internalnapi.Value { return eve.valueOf.NapiValue() }
func (eve *EventEmitter) NapiEnv() internalnapi.Env     { return eve.env.NapiValue() }

func (eve *EventEmitter) Emit(eventName string, values ...napi.ValueType) (bool, error) {
	name, err := napi.CreateString(eve.env, eventName)
	if err != nil {
		return false, err
	}
	args := append([]napi.ValueType{name}, values...)
	n1, err := eve.call("emit", args...)
	if err != nil {
		return false, err
	}
	return napi.ToBoolean(n1).Value()
}

func (eve *EventEmitter) Once(eventName string, callback *napi.Function) error {
	str, err := napi.CreateString(eve.env, eventName)
	if err != nil {
		return err
	}
	_, err = eve.call("once", str, callback)
	return err
}

func (eve *EventEmitter) On(eventName string, callback *napi.Function) error {
	str, err := napi.CreateString(eve.env, eventName)
	if err != nil {
		return err
	}
	_, err = eve.call("on", str, callback)
	return err
}

func (eve *EventEmitter) PrependOnceListener(eventName string, callback *napi.Function) error {
	str, err := napi.CreateString(eve.env, eventName)
	if err != nil {
		return err
	}
	_, err = eve.call("prependOnceListener", str, callback)
	return err
}

func (eve *EventEmitter) PrependListener(eventName string, callback *napi.Function) error {
	str, err := napi.CreateString(eve.env, eventName)
	if err != nil {
		return err
	}
	_, err = eve.call("prependListener", str, callback)
	return err
}

func (eve *EventEmitter) EventNames(_ ...string) (names []*napi.String, err error) {
	fn, err := eve.valueOf.Get("eventNames")
	if err != nil {
		return
	}

	array, err := napi.ToFunction(fn).CallWithGlobal(eve.valueOf)
	if err != nil {
		return
	}

	napiArray := napi.ToArray(array)
	size, err := napiArray.Length()
	if err != nil {
		return
	}

	names = make([]*napi.String, size)
	for index := range size {
		valueOf, err := napiArray.Get(index)
		if err != nil {
			return nil, err
		}
		names[index] = napi.ToString(valueOf)
	}

	return
}

func (eve *EventEmitter) RemoveListener(eventName string, callback *napi.Function) (err error) {
	fn, err := eve.valueOf.Get("removeListener")
	if err != nil {
		return
	}
	str, _ := napi.CreateString(eve.env, eventName)
	_, err = napi.ToFunction(fn).CallWithGlobal(eve.valueOf, str, callback)
	return
}

func (eve *EventEmitter) RemoveAllListeners(eventName ...string) error {
	var args []napi.ValueType
	for _, name := range eventName {
		value, err := napi.CreateString(eve.env, name)
		if err != nil {
			return err
		}
		args = append(args, value)
	}
	_, err := eve.call("removeAllListeners", args...)
	return err
}

func (eve *EventEmitter) GetMaxListeners() (max int64, err error) {
	fn, err := eve.valueOf.Get("getMaxListeners")
	if err != nil {
		return
	}
	numbers, err := napi.ToFunction(fn).CallWithGlobal(eve.valueOf)
	if err == nil {
		max, err = napi.ToNumber(numbers).Int()
	}
	return
}

func (eve *EventEmitter) SetMaxListeners(size int64) (err error) {
	fn, err := eve.valueOf.Get("setMaxListeners")
	if err != nil {
		return
	}
	v, _ := napi.CreateNumber(eve.env, size)
	_, err = napi.ToFunction(fn).CallWithGlobal(eve.valueOf, v)
	return
}

func (eve *EventEmitter) ListenerCount(eventName string, callback ...*napi.Function) (count int64, err error) {
	fn, err := eve.valueOf.Get("listenerCount")
	if err != nil {
		return
	}

	var res napi.ValueType

	str, _ := napi.CreateString(eve.env, eventName)
	if len(callback) == 1 {
		res, err = napi.ToFunction(fn).CallWithGlobal(eve.valueOf, str, callback[0])
	} else {
		res, err = napi.ToFunction(fn).CallWithGlobal(eve.valueOf, str)
	}
	if err == nil {
		count, err = napi.ToNumber(res).Int()
	}
	return
}

func (eve *EventEmitter) Listeners(eventName string) (callbacks []*napi.Function, err error) {
	fn, err := eve.valueOf.Get("listeners")
	if err != nil {
		return
	}

	str, _ := napi.CreateString(eve.env, eventName)
	array, err := napi.ToFunction(fn).CallWithGlobal(eve.valueOf, str)
	if err != nil {
		return
	}

	napiArray := napi.ToArray(array)
	size, err := napiArray.Length()
	if err != nil {
		return
	}

	callbacks = make([]*napi.Function, size)
	for index := range size {
		valueOf, err := napiArray.Get(index)
		if err != nil {
			return nil, err
		}
		callbacks[index] = napi.ToFunction(valueOf)
	}

	return
}

// ToEventEmitter wraps an existing JavaScript EventEmitter in its current scope.
// All methods on EventEmitter must be called on the owning Node thread.
func ToEventEmitter(value napi.ValueType) *EventEmitter {
	return &EventEmitter{env: value.Env(), valueOf: napi.ToObject(value)}
}

func (eve *EventEmitter) call(method string, args ...napi.ValueType) (napi.ValueType, error) {
	fn, err := eve.valueOf.Get(method)
	if err != nil {
		return nil, err
	}
	return napi.ToFunction(fn).CallWithGlobal(eve.valueOf, args...)
}
