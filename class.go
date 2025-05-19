package napi

import (
	"fmt"
	"reflect"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type Class[T ClassType] struct {
	value
	Class T
}

type PropertyAttributes int

type PropertyDescriptor struct {
	Name       string
	Method     Callback
	Getter     Callback
	Setter     Callback
	Value      ValueType
	Attributes PropertyAttributes
}

// Base to class declaration
type ClassType interface {
	Contructor(*CallbackInfo) (*Object, error)
}

var ClassFuncMathod = reflect.TypeFor[Callback]()

const (
	PropertyAttributesWritable          = PropertyAttributes(napi.Writable)
	PropertyAttributesEnumerable        = PropertyAttributes(napi.Enumerable)
	PropertyAttributesConfigurable      = PropertyAttributes(napi.Configurable)
	PropertyAttributesStatic            = PropertyAttributes(napi.Static)
	PropertyAttributesDefault           = PropertyAttributes(napi.Default)
	PropertyAttributesDefaultMethod     = PropertyAttributes(napi.DefaultMethod)
	PropertyAttributesDefaultJSProperty = PropertyAttributes(napi.DefaultJSProperty)
)

func CreateClass[T ClassType](env EnvType) (*Class[T], error) {
	ptrType := reflect.TypeFor[T]()
	switch ptrType.Kind() {
	default:
		return nil, fmt.Errorf("type %s is not a struct", ptrType.Name())
	case reflect.Pointer:
		elem := ptrType.Elem()
		if elem.Kind() == reflect.Struct {
			return classMount[T](env, elem)
		}
		fallthrough
	case reflect.Struct:
		return classMount[T](env, ptrType)
	}
}

func classMount[T ClassType](env EnvType, ptr reflect.Type) (*Class[T], error) {
	valueOf := reflect.New(ptr)

	var propertys []*PropertyDescriptor
	for methodIndex := range ptr.NumMethod() {
		method := ptr.Method(methodIndex)
		if method.Type.Kind() != reflect.Func || !method.IsExported() {
			continue
		} else if !method.Type.Implements(ClassFuncMathod) {
			continue
		} else if method.Name == "Contructor" {
			continue
		}

		propertys = append(propertys, &PropertyDescriptor{
			Name:       method.Name,
			Attributes: PropertyAttributes(PropertyAttributesStatic),
			Method:     valueOf.Method(methodIndex).Interface().(Callback),
			Getter:     nil,
			Setter:     nil,
			Value:      nil,
		})
	}

	var napiAtr []napi.PropertyDescriptor
	for _, value := range propertys {
		name, err := CreateString(env, value.Name)
		if err != nil {
			return nil, err
		}

		method, err := CreateFunction(env, value.Name, value.Method)
		if err != nil {
			return nil, err
		}

		napiAtr = append(napiAtr, napi.PropertyDescriptor{
			Utf8name:   value.Name,
			Name:       name.NapiValue(),
			Method:     method.NapiCallback(),
			Attributes: napi.PropertyAttributes(value.Attributes),
		})
	}

	jsConstruct, err := CreateFunction(env, "constructor", func(ci *CallbackInfo) (ValueType, error) {
		res := valueOf.MethodByName("Contructor").Call([]reflect.Value{reflect.ValueOf(ci)})
		obj, err := res[0].Interface().(*Object), res[1].Interface().(error)
		return N_APIValue(obj.Env(), obj.NapiValue()), err
	})
	if err != nil {
		return nil, err
	}

	jsValue, status := napi.DefineClass(env.NapiValue(), ptr.Name(),
		jsConstruct.fn,
		napiAtr,
	)

	if err := status.ToError(); err != nil {
		return nil, err
	}

	return &Class[T]{
		value: N_APIValue(env, jsValue),
		Class: valueOf.Interface().(T),
	}, nil
}
