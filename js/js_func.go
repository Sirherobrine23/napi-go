package js

import (
	"fmt"
	"reflect"
	"runtime"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	internalNapi "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

// Create napi bind to golang functions
func GoFuncOf(env napi.EnvType, function any) (napi.ValueType, error) {
	return funcOf(env, reflect.ValueOf(function))
}

func funcOf(env napi.EnvType, ptr reflect.Value) (napi.ValueType, error) {
	if ptr.Kind() != reflect.Func {
		return nil, fmt.Errorf("return function to return napi value")
	}

	funcName := runtime.FuncForPC(ptr.Pointer()).Name()
	switch v := ptr.Interface().(type) {
	case nil:
		return nil, nil
	case napi.Callback:
		return napi.CreateFunction(env, funcName, v)
	case internalNapi.Callback:
		return napi.CreateFunctionNapi(env, funcName, v)
	default:
		return napi.CreateFunction(env, funcName, func(env napi.EnvType, this napi.ValueType, args []napi.ValueType) (napi.ValueType, error) {
			fnType := ptr.Type()
			returnValues := []reflect.Value{}
			switch {
			case fnType.NumIn() == 0 && fnType.NumOut() == 0:
				returnValues = ptr.Call([]reflect.Value{})
			case !fnType.IsVariadic():
				returnValues = ptr.Call(goValuesInFunc(ptr, args, false))
			default:
				returnValues = ptr.CallSlice(goValuesInFunc(ptr, args, true))
			}

			switch len(returnValues) {
			case 0:
				return nil, nil
			case 1:
				return valueOf(env, returnValues[0])
			default:
				lastValue := returnValues[len(returnValues)-1]
				if lastValue.CanConvert(reflect.TypeFor[error]()) {
					returnValues = returnValues[:len(returnValues)-1]
					if !lastValue.IsNil() {
						if err := lastValue.Interface().(error); err != nil {
							return nil, err
						}
					}

				}

				switch len(returnValues) {
				case 1:
					// Return value from return
					return valueOf(env, returnValues[0])
				default:
					// Create array
					arr, err := napi.CreateArray(env, len(returnValues))
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
			}
		})
	}
}

func goValuesInFunc(ptr reflect.Value, jsArgs []napi.ValueType, variadic bool) (values []reflect.Value) {
	if variadic && (ptr.Type().NumIn()-1 > 0) && ptr.Type().NumIn()-1 < len(jsArgs) {
		panic(fmt.Errorf("require minimun %d arguments, called with %d", ptr.Type().NumIn()-1, len(jsArgs)))
	} else if !variadic &&ptr.Type().NumIn() != len(jsArgs) {
		panic(fmt.Errorf("require %d arguments, called with %d", ptr.Type().NumIn(), len(jsArgs)))
	}

	size := ptr.Type().NumIn()
	if variadic {
		size--
	}

	values = make([]reflect.Value, size)
	for index := range values {
		valueOf := reflect.New(ptr.Type().In(index))
		if err := valueFrom(jsArgs[index], valueOf); err != nil {
			panic(err)
		}
		values[index] = valueOf
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
