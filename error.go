package napi

import (
	"runtime"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type Error struct{ value }

func ToError(o ValueType) *Error { return &Error{o} }

// CreateError creates a new JavaScript Error object in the given N-API environment with the specified message.
// It returns a pointer to the Error object and an error if the creation fails at any step.
func CreateError(env EnvType, msg string) (*Error, error) {
	napiMsg, err := CreateString(env, msg)
	if err != nil {
		return nil, err
	}
	napiValue, err := mustValueErr(napi.CreateError(env.NapiValue(), nil, napiMsg.NapiValue()))
	if err != nil {
		return nil, err
	}
	return ToError(N_APIValue(env, napiValue)), nil
}

// ThrowAsJavaScriptException throws the current Error as a JavaScript exception
// in the associated N-API environment. It returns an error if the operation fails.
func (er *Error) ThrowAsJavaScriptException() error {
	return singleMustValueErr(napi.Throw(er.NapiEnv(), er.NapiValue()))
}

// ThrowError throws a JavaScript error in the given N-API environment with the specified code and error message.
// If the code is an empty string, it captures the current Go stack trace and uses it as the error code.
// Returns an error if the underlying N-API call fails.
//
// Parameters:
//   - env: The N-API environment in which to throw the error.
//   - code: The error code to associate with the thrown error. If empty, the current stack trace is used.
//   - err: The error message to be thrown.
func ThrowError(env EnvType, code, err string) error {
	if code == "" {
		stackTraceBuf := make([]byte, 8192)
		stackTraceSz := runtime.Stack(stackTraceBuf, false)
		code = string(stackTraceBuf[:stackTraceSz])
	}
	return singleMustValueErr(napi.ThrowError(env.NapiValue(), code, err))
}
