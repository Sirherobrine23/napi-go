package napi

import (
	"fmt"
	"reflect"
	"unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type (
	PropertyDescriptor = napi.PropertyDescriptor
	
	ClassType interface {
		Contructor(env EnvType, this ValueType, args []ValueType) (*Object, error)
	}

	Class[T ClassType] struct {
		value
		Class        T
		classReflect reflect.Type
	}
)

const (
	PropertyAttributesDefault           = napi.Default
	PropertyAttributesWritable          = napi.Writable
	PropertyAttributesEnumerable        = napi.Enumerable
	PropertyAttributesConfigurable      = napi.Configurable
	PropertyAttributesStatic            = napi.Static
	PropertyAttributesDefaultMethod     = napi.DefaultMethod
	PropertyAttributesDefaultJSProperty = napi.DefaultJSProperty
)

func CreateClass[T ClassType](env EnvType, propertys []PropertyDescriptor) (*Class[T], error) {
	ptrType := reflect.TypeFor[T]()
	switch ptrType.Kind() {
	case reflect.Struct:
	case reflect.Pointer:
		if ptrType = ptrType.Elem(); ptrType.Kind() != reflect.Struct {
			return nil, fmt.Errorf("type %s is not a struct", ptrType.Name())
		}
	default:
		return nil, fmt.Errorf("type %s is not a struct", ptrType.Name())
	}

	startStruct := &Class[T]{classReflect: ptrType}
	value, status := napi.DefineClass(env.NapiValue(), ptrType.Name(),
		func(env napi.Env, info napi.CallbackInfo) napi.Value {
			startStruct.Class = reflect.New(ptrType).Interface().(T)
			cbInfo, status := napi.GetCbInfo(env, info)
			if err := status.ToError(); err != nil {
				ThrowError(N_APIEnv(env), "", err.Error())
				return nil
			}

			gonapiEnv := N_APIEnv(env)
			this := N_APIValue(gonapiEnv, cbInfo.This)
			args := make([]ValueType, len(cbInfo.Args))
			for i, cbArg := range cbInfo.Args {
				args[i] = N_APIValue(gonapiEnv, cbArg)
			}

			res, err := startStruct.Class.Contructor(gonapiEnv, this, args)
			if err != nil {
				ThrowError(gonapiEnv, "", err.Error())
			} else if res == nil {
				und, _ := gonapiEnv.Undefined()
				return und.NapiValue()
			}
			napi.Wrap(env, res.NapiValue(), unsafe.Pointer(startStruct), nil, nil)
			return res.NapiValue()
		}, propertys)
	if err := status.ToError(); err != nil {
		return nil, err
	}
	startStruct.value = N_APIValue(env, value)
	return startStruct, nil
}
