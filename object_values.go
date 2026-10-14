package napi

import (
	"fmt"
	"slices"

	internal "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

// ObjectOf converts a Go map or struct, or wraps an existing JavaScript object.
// Arrays, functions and other JavaScript objects are accepted; primitives and
// null are rejected. Like every Object operation, it runs on the owning Node
// thread and the returned handle belongs to the current handle scope.
func ObjectOf(env EnvType, input any) (*Object, error) {
	value, err := ValueOf(env, input)
	if err != nil {
		return nil, err
	}
	kind, status := internal.Typeof(env.NapiValue(), value.NapiValue())
	if err := status.ToError(); err != nil {
		return nil, err
	}
	if kind != internal.ValueTypeObject && kind != internal.ValueTypeFunction {
		return nil, fmt.Errorf("ObjectOf requires an object, got %s", kind)
	}
	return ToObject(value), nil
}

// SetValue converts a Go value using ValueOf and assigns a named property.
// Existing ValueType handles must belong to this object's environment.
// It retains Object.Set semantics: a non-writable property may ignore the write.
func (obj *Object) SetValue(key string, input any) error {
	value, err := ValueOf(obj.Env(), input)
	if err != nil {
		return fmt.Errorf("property %q: %w", key, err)
	}
	if err := obj.Set(key, value); err != nil {
		return fmt.Errorf("property %q: %w", key, err)
	}
	return nil
}

// Assign sets Go values in sorted key order. Assignment uses JavaScript setters
// and is not atomic: properties preceding a failed assignment remain changed.
func (obj *Object) Assign(properties map[string]any) error {
	keys := make([]string, 0, len(properties))
	for key := range properties {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		if err := obj.SetValue(key, properties[key]); err != nil {
			return err
		}
	}
	return nil
}

// GetInto decodes a property into a non-nil pointer, without JavaScript coercion.
func (obj *Object) GetInto(key string, target any) error {
	value, err := obj.Get(key)
	if err == nil {
		err = ValueFrom(value, target)
	}
	if err != nil {
		return fmt.Errorf("property %q: %w", key, err)
	}
	return nil
}

// GetAs is the typed form of GetInto. Go methods cannot have type parameters.
func GetAs[T any](obj *Object, key string) (T, error) {
	var result T
	err := obj.GetInto(key, &result)
	return result, err
}

// SetMethod registers a Callback or an ordinary Go function as a property.
// Ordinary Go functions use GoFuncOf argument/result conversion; use Callback
// when the implementation needs CallbackInfo.This or the Node environment.
func (obj *Object) SetMethod(name string, method any) error {
	function, err := GoFuncOf(obj.Env(), method)
	if err != nil {
		return fmt.Errorf("method %q: %w", name, err)
	}
	return obj.Set(name, function)
}

// CallMethod converts Go arguments and calls obj[name] with obj as JavaScript
// this. It looks up the method on each call and rejects non-function properties.
func (obj *Object) CallMethod(name string, args ...any) (ValueType, error) {
	method, err := obj.Get(name)
	if err != nil {
		return nil, err
	}
	kind, status := internal.Typeof(obj.NapiEnv(), method.NapiValue())
	if err := status.ToError(); err != nil {
		return nil, err
	}
	if kind != internal.ValueTypeFunction {
		return nil, fmt.Errorf("property %q is %s, expected function", name, kind)
	}
	return ToFunction(method).CallWithThis(obj, args...)
}
