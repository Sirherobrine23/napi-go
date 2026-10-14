package napi

import (
	"fmt"
	"strings"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type PropertyAttributes = napi.PropertyAttributes

const (
	Default           PropertyAttributes = napi.Default
	Writable          PropertyAttributes = napi.Writable
	Enumerable        PropertyAttributes = napi.Enumerable
	Configurable      PropertyAttributes = napi.Configurable
	Static            PropertyAttributes = napi.Static
	DefaultMethod     PropertyAttributes = napi.DefaultMethod
	DefaultJSProperty PropertyAttributes = napi.DefaultJSProperty
)

// PropertyDescriptor defines a named or symbol-keyed value, method or accessor.
// Name takes precedence over Utf8name. Static places a class property on the
// constructor; otherwise it is placed on the prototype. Attributes default to 0.
// Descriptor callbacks are retained until their Node environment is destroyed.
type PropertyDescriptor struct {
	Utf8name   string
	Name       ValueType
	Method     Callback
	Getter     Callback
	Setter     Callback
	Value      ValueType
	Attributes PropertyAttributes
}

// Class is the JavaScript constructor returned by napi_define_class. Like other
// ValueTypes it is valid only in its originating Node thread and handle scope.
type Class struct{ *Function }

func ToClass(value ValueType) *Class { return &Class{ToFunction(value)} }

// CreateClass defines a class with native constructor, methods and accessors.
// The constructor must wrap/initialize CallbackInfo.This and may return nil to
// use that object. Calling the constructor without new throws a TypeError.
func CreateClass(env EnvType, name string, constructor Callback, properties ...PropertyDescriptor) (*Class, error) {
	if constructor == nil {
		return nil, fmt.Errorf("constructor cannot be nil")
	}
	descriptors, err := propertiesToNapi(env, properties)
	if err != nil {
		return nil, err
	}
	callback := callbackToNapi(func(info *CallbackInfo) (ValueType, error) {
		target, err := info.NewTarget()
		if err != nil {
			return nil, err
		}
		if target == nil {
			napi.ThrowTypeError(env.NapiValue(), "", "class constructor must be called with new")
			return nil, nil
		}
		return constructor(info)
	})
	value, status := napi.DefineClass(env.NapiValue(), name, callback, descriptors)
	if err := status.ToError(); err != nil {
		return nil, err
	}
	class := ToClass(N_APIValue(env, value))
	class.fn = callback
	return class, nil
}

// New creates a JavaScript instance on the Node thread, with zero or more args.
func (class *Class) New(args ...ValueType) (*Object, error) {
	argv := make([]napi.Value, len(args))
	for i, arg := range args {
		if arg == nil || arg.NapiEnv() != class.NapiEnv() {
			return nil, fmt.Errorf("argument %d must belong to the class environment", i)
		}
		argv[i] = arg.NapiValue()
	}
	value, status := napi.NewInstance(class.NapiEnv(), class.NapiValue(), len(argv), argv)
	if err := status.ToError(); err != nil {
		return nil, err
	}
	return ToObject(N_APIValue(class.Env(), value)), nil
}

func propertiesToNapi(env EnvType, properties []PropertyDescriptor) ([]napi.PropertyDescriptor, error) {
	descriptors := make([]napi.PropertyDescriptor, len(properties))
	for i, property := range properties {
		if property.Name == nil && strings.ContainsRune(property.Utf8name, 0) {
			return nil, fmt.Errorf("property %d contains a NUL byte", i)
		}
		accessor := property.Getter != nil || property.Setter != nil
		if (property.Method != nil && (accessor || property.Value != nil)) || (accessor && property.Value != nil) {
			return nil, fmt.Errorf("property %d must be a value, method or accessor", i)
		}
		p := &descriptors[i]
		p.Utf8name, p.Attributes = property.Utf8name, property.Attributes
		for _, value := range []ValueType{property.Name, property.Value} {
			if value != nil && value.NapiEnv() != env.NapiValue() {
				return nil, fmt.Errorf("property %d belongs to another environment", i)
			}
		}
		if property.Name != nil {
			p.Name = property.Name.NapiValue()
		}
		if property.Value != nil {
			p.Value = property.Value.NapiValue()
		}
		if property.Method != nil {
			p.Method = callbackToNapi(property.Method)
		}
		if property.Getter != nil {
			p.Getter = callbackToNapi(property.Getter)
		}
		if property.Setter != nil {
			p.Setter = callbackToNapi(property.Setter)
		}
	}
	return descriptors, nil
}

// DefineProperties defines native methods, accessors and data on this object.
func (obj *Object) DefineProperties(properties ...PropertyDescriptor) error {
	descriptors, err := propertiesToNapi(obj.Env(), properties)
	if err != nil {
		return err
	}
	return napi.DefineProperties(obj.NapiEnv(), obj.NapiValue(), descriptors).ToError()
}

// FinalizeCallback releases Go-owned instance resources during GC or environment
// teardown. It must not invoke JavaScript or Node-API, and panics are contained.
type FinalizeCallback func(data any)

// Wrap attaches Go state without passing a Go pointer to C. Each object may have
// one wrap. The optional finalizer runs once unless RemoveWrap detaches the state.
func (obj *Object) Wrap(data any, finalize FinalizeCallback) error {
	return napi.WrapGoValue(obj.NapiEnv(), obj.NapiValue(), data, napi.WrapFinalize(finalize)).ToError()
}

func (obj *Object) Unwrap() (any, error) {
	data, status := napi.UnwrapGoValue(obj.NapiEnv(), obj.NapiValue(), false)
	return data, status.ToError()
}

// RemoveWrap transfers ownership of state to the caller and cancels its finalizer.
func (obj *Object) RemoveWrap() (any, error) {
	data, status := napi.UnwrapGoValue(obj.NapiEnv(), obj.NapiValue(), true)
	return data, status.ToError()
}
