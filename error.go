package napi

import (
	"runtime"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type Error struct{ value }

func ToError(o ValueType) *Error { return &Error{o} }

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

func (er *Error) ThrowAsJavaScriptException() error {
	return singleMustValueErr(napi.Throw(er.NapiEnv(), er.NapiValue()))
}

// This throws a JavaScript Error with the text provided.
func ThrowError(env EnvType, code, err string) error {
	if code == "" {
		stackTraceBuf := make([]byte, 8192)
		stackTraceSz := runtime.Stack(stackTraceBuf, false)
		code = string(stackTraceBuf[:stackTraceSz])
	}
	return singleMustValueErr(napi.ThrowError(env.NapiValue(), code, err))
}
