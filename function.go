package napi

import (
	"fmt"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type Function struct {
	value
	fn napi.Callback
}

// Function to call on Javascript caller
type Callback func(*CallbackInfo) (ValueType, error)

// Values from [napi.CallbackInfo]
type CallbackInfo struct {
	Env  EnvType
	This ValueType
	Args []ValueType

	info napi.CallbackInfo
}

func (call *CallbackInfo) NewTarget() (ValueType, error) {
	v, status := napi.GetNewTarget(call.Env.NapiValue(), call.info)
	if err := status.ToError(); err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	return N_APIValue(call.Env, v), nil
}

// Convert [ValueType] to [*Function]
func ToFunction(o ValueType) *Function { return &Function{o, nil} }

// CreateFunction creates a new JavaScript function in the given N-API environment with the specified name and callback.
// The callback is invoked when the JavaScript function is called, receiving a CallbackInfo containing the environment,
// 'this' value, arguments, and callback info. Any errors returned by the callback or panics are converted to JavaScript
// exceptions. If the callback returns nil, the JavaScript 'undefined' value is returned. If the callback returns a value
// of TypeError, it is thrown as a JavaScript exception.
func CreateFunction(env EnvType, name string, callback Callback) (*Function, error) {
	if callback == nil {
		return nil, fmt.Errorf("callback cannot be nil")
	}
	return CreateFunctionNapi(env, name, callbackToNapi(callback))
}

func callbackToNapi(callback Callback) napi.Callback {
	return func(napiEnv napi.Env, info napi.CallbackInfo) napi.Value {
		env := N_APIEnv(napiEnv)
		cbInfo, status := napi.GetCbInfo(napiEnv, info)
		if err := status.ToError(); err != nil {
			ThrowError(env, "", err.Error())
			return nil
		}

		this := N_APIValue(env, cbInfo.This)
		args := make([]ValueType, len(cbInfo.Args))
		for i, cbArg := range cbInfo.Args {
			args[i] = N_APIValue(env, cbArg)
		}

		defer func() {
			if err := recover(); err != nil {
				switch v := err.(type) {
				case error:
					ThrowError(env, "", v.Error())
				default:
					ThrowError(env, "", fmt.Sprintf("panic recover: %v", err))
				}
			}
		}()

		res, err := callback(&CallbackInfo{env, this, args, info})
		switch {
		case err != nil:
			ThrowError(env, "", err.Error())
			return nil
		case res == nil:
			und, _ := env.Undefined()
			return und.NapiValue()
		default:
			typeOf, _ := res.Type()
			if typeOf == TypeError {
				ToError(res).ThrowAsJavaScriptException()
				return nil
			}
			return res.NapiValue()
		}
	}
}

// Create function from internal [napi.Callback]
func CreateFunctionNapi(env EnvType, name string, callback napi.Callback) (*Function, error) {
	fnCall, err := napi.CreateFunction(env.NapiValue(), name, callback)
	if err := err.ToError(); err != nil {
		return nil, err
	}
	fn := ToFunction(N_APIValue(env, fnCall))
	fn.fn = callback
	return fn, nil
}

func (fn *Function) NapiCallback() napi.Callback {
	return fn.fn
}

func (fn *Function) internalCall(this napi.Value, argc int, argv []napi.Value) (ValueType, error) {
	// napi_call_function(env, global, add_two, argc, argv, &return_val);
	res, err := napi.CallFunction(fn.NapiEnv(), this, fn.NapiValue(), argc, argv)
	if err := err.ToError(); err != nil {
		return nil, err
	}
	return N_APIValue(fn.Env(), res), nil
}

// Call function with custom global/this value
func (fn *Function) CallWithGlobal(this ValueType, args ...ValueType) (ValueType, error) {
	argv := make([]napi.Value, len(args))
	for index, valueof := range args {
		argv[index] = valueof.NapiValue()
	}
	return fn.internalCall(this.NapiValue(), len(args), argv)
}

// Call function with args
func (fn *Function) Call(args ...ValueType) (ValueType, error) {
	global, err := fn.Env().Global()
	if err != nil {
		return nil, err
	}
	return fn.CallWithGlobal(global, args...)
}
