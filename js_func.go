package napi

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	internalNapi "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

var typeofError = reflect.TypeFor[error]()

// GoFuncOf wraps a Go function as a JavaScript-compatible function for use with the given environment.
// It takes an EnvType representing the JavaScript environment and a Go function (of any type).
// Returns a ValueType representing the JavaScript function and an error if the wrapping fails.
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
			var goFnReturn []reflect.Value
			in, out, veridict := ptr.Type().NumIn(), ptr.Type().NumOut(), ptr.Type().IsVariadic()

			// Check to call
			switch {
			case in == 0 && out == 0: // only call
				goFnReturn = ptr.Call([]reflect.Value{})
			case !veridict: // call same args
				goFnReturn = ptr.Call(goValuesInFunc(ptr, args, false))
			default: // call with slice on end
				goFnReturn = ptr.CallSlice(goValuesInFunc(ptr, args, true))
			}

			// Check for last element is error
			if len(goFnReturn) > 0 {
				lastValue := goFnReturn[len(goFnReturn)-1]
				if lastValue.CanConvert(typeofError) {
					goFnReturn = goFnReturn[:len(goFnReturn)-1] // remove last element from return
					if !lastValue.IsNil() {                           // check if not is nil to throw error in javascript
						return nil, lastValue.Interface().(error)
					}
				}
			}

			// Check return value
			switch len(goFnReturn) {
			case 0: // not value to return
				return env.Undefined()
			case 1: // Check if error or value to return
				return valueOf(env, goFnReturn[0])
			}

			// Convert to array return and check if latest is error
			napiValueReturn, err := CreateArray(env, len(goFnReturn))
			if err != nil {
				return nil, err
			}

			// Append values to js array
			for index, value := range goFnReturn {
				napiValue, err := valueOf(env, value)
				if err != nil {
					return nil, err
				} else if err = napiValueReturn.Set(index, napiValue); err != nil {
					return nil, err
				}
			}
			return napiValueReturn, nil
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
