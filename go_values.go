package napi

import (
	"fmt"
	"reflect"
)

// Decode converts a JavaScript value to T using ValueFrom, without coercion.
func Decode[T any](value ValueType) (T, error) {
	var result T
	err := ValueFrom(value, &result)
	return result, err
}

// ArgInto decodes one callback argument and reports a missing argument as an
// error. Defaults and optional arguments remain explicit in the callback.
func (call *CallbackInfo) ArgInto(index int, target any) error {
	if index < 0 || index >= len(call.Args) {
		return fmt.Errorf("argument %d is missing (received %d)", index, len(call.Args))
	}
	if err := ValueFrom(call.Args[index], target); err != nil {
		return fmt.Errorf("argument %d: %w", index, err)
	}
	return nil
}

// ArgAs is the typed form of CallbackInfo.ArgInto.
func ArgAs[T any](call *CallbackInfo, index int) (T, error) {
	var result T
	err := call.ArgInto(index, &result)
	return result, err
}

// CallValues calls a function with Go arguments and the global object as this.
func (fn *Function) CallValues(args ...any) (ValueType, error) {
	values, err := valuesOf(fn.Env(), args)
	if err != nil {
		return nil, err
	}
	return fn.Call(values...)
}

// CallWithThis calls a function with an explicit receiver and Go arguments.
func (fn *Function) CallWithThis(this ValueType, args ...any) (ValueType, error) {
	values, err := valuesOf(fn.Env(), args)
	if err != nil {
		return nil, err
	}
	return fn.CallWithGlobal(this, values...)
}

// NewValues creates a class instance with Go arguments on the Node thread.
func (class *Class) NewValues(args ...any) (*Object, error) {
	values, err := valuesOf(class.Env(), args)
	if err != nil {
		return nil, err
	}
	return class.New(values...)
}

// UnwrapAs retrieves wrapped Go state and checks its type instead of requiring
// an unchecked assertion. It does not convert or copy the stored state.
func UnwrapAs[T any](obj *Object) (T, error) {
	var zero T
	state, err := obj.Unwrap()
	if err != nil {
		return zero, err
	}
	result, ok := state.(T)
	if !ok {
		return zero, fmt.Errorf("wrapped state is %T, expected %s", state, reflect.TypeFor[T]())
	}
	return result, nil
}

func valuesOf(env EnvType, args []any) ([]ValueType, error) {
	values := make([]ValueType, len(args))
	for index, arg := range args {
		value, err := ValueOf(env, arg)
		if err != nil {
			return nil, fmt.Errorf("argument %d: %w", index, err)
		}
		values[index] = value
	}
	return values, nil
}

func nilValue(value any) bool {
	if value == nil {
		return true
	}
	ptr := reflect.ValueOf(value)
	switch ptr.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Func, reflect.Map, reflect.Slice, reflect.Chan:
		return ptr.IsNil()
	}
	return false
}

func validateValue(env EnvType, value ValueType) error {
	if nilValue(value) || value.NapiValue() == nil {
		return fmt.Errorf("JavaScript value is nil")
	}
	if value.NapiEnv() != env.NapiValue() {
		return fmt.Errorf("JavaScript value belongs to another environment")
	}
	return nil
}
