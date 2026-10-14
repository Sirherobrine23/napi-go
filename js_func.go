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
	if !ptr.IsValid() || ptr.Kind() != reflect.Func {
		return nil, fmt.Errorf("GoFuncOf requires a Go function")
	} else if !ptr.CanInterface() {
		return nil, fmt.Errorf("cannot check function type")
	} else if ptr.IsNil() {
		return nil, fmt.Errorf("GoFuncOf requires a non-nil function")
	}

	funcName := strings.ReplaceAll(runtime.FuncForPC(ptr.Pointer()).Name(), ".", "_")
	if ptr.Type().ConvertibleTo(reflect.TypeFor[Callback]()) {
		return CreateFunction(env, funcName, ptr.Convert(reflect.TypeFor[Callback]()).Interface().(Callback))
	}
	switch v := ptr.Interface().(type) {
	case Callback: // return function value
		return CreateFunction(env, funcName, v)
	case internalNapi.Callback: // return internal/napi function value
		return CreateFunctionNapi(env, funcName, v)
	default: // Convert go function to javascript function
		return CreateFunction(env, funcName, func(ci *CallbackInfo) (ValueType, error) {
			var goFnReturn []reflect.Value
			variadic := ptr.Type().IsVariadic()

			// Check to call
			switch {
			case !variadic: // call same args
				goFnReturn = ptr.Call(goValuesInFunc(ptr, ci.Args, false))
			default: // call with slice on end
				goFnReturn = ptr.CallSlice(goValuesInFunc(ptr, ci.Args, true))
			}

			// Check for last element is error
			if len(goFnReturn) > 0 {
				lastValue := goFnReturn[len(goFnReturn)-1]
				if lastValue.Type().Implements(typeofError) {
					goFnReturn = goFnReturn[:len(goFnReturn)-1] // remove last element from return
					if !nilValue(lastValue.Interface()) {
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
	if variadic && len(jsArgs) < ptr.Type().NumIn()-1 {
		panic(fmt.Errorf("require minimum %d arguments, called with %d", ptr.Type().NumIn()-1, len(jsArgs)))
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
		values[index] = reflect.New(ptrType).Elem()
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
