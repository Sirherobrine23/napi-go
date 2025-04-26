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
		value, err := internalNapi.MustValueErr(internalNapi.CreateFunction(env.NapiValue(), funcName, v))
		if err != nil {
			return nil, err
		}
		return napi.FunctionFromValue(env, value), nil
	default:
		return napi.CreateFunction(env, funcName, func(env napi.EnvType, this napi.ValueType, args []napi.ValueType) (napi.ValueType, error) {
			fnType := ptr.Type()

			// Go => Node
			// func() => function(): void
			// func(value1, value2) => function(value1, value2): void
			//
			// func() error => function(): throw Error
			// func(value1, value2) error => function(value1, value2): throw Error
			// func(value1, value2) (any, error) => function(value1, value2): (Array or object||throw Error)
			//
			// func(value1, value2) (any, any2...) => function(value1, value2): Array
			// func(value1, value2) (any, any2..., error) => function(value1, value2): (Array||throw Error)

			// call function
			returnValues, ins := []reflect.Value{}, []reflect.Value{}
			if fnType.IsVariadic() {
				if fnType.NumIn() > 0 {
					lastElement := fnType.NumIn() - 1
					for i := range lastElement {
						point := reflect.New(fnType.In(i)).Elem()
						napiValue := args[i]
						if err := reflectFromTypedValue(env, point, napiValue); err != nil {
							return nil, err
						}
						ins = append(ins, point)
					}

					args := args[lastElement:]
					point := reflect.MakeSlice(fnType.In(lastElement), len(args), len(args))
					for index := range args {
						if err := reflectFromTypedValue(env, point.Index(index), args[index]); err != nil {
							return nil, err
						}
					}
					ins = append(ins, point)
				}
				returnValues = ptr.CallSlice(ins)
			} else {
				if fnType.NumIn() > len(args) {
					for i := range fnType.NumIn() {
						point := reflect.New(fnType.In(i)).Elem()
						if len(args) <= i {
							ins = append(ins, point)
							continue
						}
						napiValue := args[i]
						if err := reflectFromTypedValue(env, point, napiValue); err != nil {
							return nil, err
						}
						ins = append(ins, point)
					}
				} else {
					for i := range fnType.NumIn() {
						point := reflect.New(fnType.In(i)).Elem()
						if len(args) <= i {
							ins = append(ins, point)
							break
						}
						napiValue := args[i]
						if err := reflectFromTypedValue(env, point, napiValue); err != nil {
							return nil, err
						}
						ins = append(ins, point)
					}
				}
				
				returnValues = ptr.Call(ins)
			}

			if len(returnValues) == 0 {
				return nil, nil
			} else if len(returnValues) == 1 {
				return ValueOf(env, returnValues[0].Interface())
			}

			arr, err := napi.CreateArrayWithLength(env, len(returnValues))
			if err != nil {
				return nil, err
			}

			if v, ok := returnValues[len(returnValues)-1].Interface().(error); ok {
				returnValues = returnValues[:len(returnValues)-1]
				if v != nil {
					return ValueOf(env, v)
				}
			}

			for index := range returnValues {
				value, err := ValueOf(env, returnValues[index].Interface())
				if err != nil {
					return nil, err
				} else if err = arr.Set(index, value); err != nil {
					return nil, err
				}
			}
			return arr, nil
		})
	}
}
