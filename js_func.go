package napi

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	internalNapi "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

// Create napi bind to golang functions
func GoFuncOf(env EnvType, function any) (ValueType, error) {
	return funcOf(env, reflect.ValueOf(function))
}

func funcOf(env EnvType, ptr reflect.Value) (ValueType, error) {
	if ptr.Kind() != reflect.Func {
		return nil, fmt.Errorf("return function to return napi value")
	} else if !ptr.IsValid() {
		return nil, fmt.Errorf("return valid reflect")
	} else if !ptr.CanInterface() {
		return nil, fmt.Errorf("cannot check function type")
	} else if ptr.IsNil() {
		// return nil, fmt.Errorf("return function, is nil")
		return nil, nil
	}

	funcName := strings.ReplaceAll(runtime.FuncForPC(ptr.Pointer()).Name(), ".", "_")
	switch v := ptr.Interface().(type) {
	case Callback: // return function value
		return CreateFunction(env, funcName, v)
	case internalNapi.Callback: // return internal/napi function value
		return CreateFunctionNapi(env, funcName, v)
	default: // Convert go function to javascript function
		return CreateFunction(env, funcName, func(env EnvType, this ValueType, args []ValueType) (ValueType, error) {
			fnType := ptr.Type()
			returnValues := []reflect.Value{}
			switch {
			case fnType.NumIn() == 0 && fnType.NumOut() == 0: // only call
				returnValues = ptr.Call([]reflect.Value{})
			case !fnType.IsVariadic(): // call same args
				returnValues = ptr.Call(goValuesInFunc(ptr, args, false))
			default: // call with slice on end
				returnValues = ptr.CallSlice(goValuesInFunc(ptr, args, true))
			}

			// Check return value
		retry:
			switch len(returnValues) {
			case 0: // not value to return
				return env.Undefined()
			case 1: // Check if error or value to return
				return valueOf(env, returnValues[0])
			default: // Convert to array return and check if latest is error
				lastValue := returnValues[len(returnValues)-1]
				if lastValue.CanConvert(reflect.TypeFor[error]()) {
					returnValues = returnValues[:len(returnValues)-1]
					if !lastValue.IsNil() {
						if err := lastValue.Interface().(error); err != nil {
							return nil, err
						}
					}
					goto retry
				}

				// Create array
				arr, err := CreateArray(env, len(returnValues))
				if err != nil {
					return nil, err
				}

				// Append values to js array
				for index, value := range returnValues {
					napiValue, err := valueOf(env, value)
					if err != nil {
						return nil, err
					} else if err = arr.Set(index, napiValue); err != nil {
						return nil, err
					}
				}
				return arr, nil
			}
		})
	}
}

// Create call value to go function
func goValuesInFunc(ptr reflect.Value, jsArgs []ValueType, variadic bool) (values []reflect.Value) {
	if variadic && (ptr.Type().NumIn()-1 > 0) && ptr.Type().NumIn()-1 < len(jsArgs) {
		panic(fmt.Errorf("require minimun %d arguments, called with %d", ptr.Type().NumIn()-1, len(jsArgs)))
	} else if !variadic && ptr.Type().NumIn() != len(jsArgs) {
		panic(fmt.Errorf("require %d arguments, called with %d", ptr.Type().NumIn(), len(jsArgs)))
	}

	size := ptr.Type().NumIn()
	if variadic {
		size-- // remove latest value to slice
	}

	// Convert value
	values = make([]reflect.Value, size)
	for index := range values {
		// Create value to append go value
		ptrType := ptr.Type().In(index)
		switch ptrType.Kind() {
		case reflect.Pointer:
			values[index] = reflect.New(ptrType.Elem())
		case reflect.Slice:
			values[index] = reflect.MakeSlice(ptrType, 0, 0)
		default:
			values[index] = reflect.New(ptrType).Elem()
		}
		if err := valueFrom(jsArgs[index], values[index]); err != nil {
			panic(err)
		}
	}

	if variadic {
		variadicType := ptr.Type().In(size).Elem()

		valueAppend := jsArgs[size:]
		valueOf := reflect.MakeSlice(reflect.SliceOf(variadicType), len(valueAppend), len(valueAppend))
		for index := range valueAppend {
			if err := valueFrom(valueAppend[index], valueOf.Index(index)); err != nil {
				panic(err)
			}
		}
		values = append(values, valueOf)
	}

	return
}
