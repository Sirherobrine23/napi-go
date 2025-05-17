package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

// EnvType defines an interface for interacting with a NAPI environment.
// It provides methods to retrieve the underlying NAPI environment, access the global object,
// and obtain the JavaScript 'undefined' and 'null' values as Go types.
type EnvType interface {
	NapiValue() napi.Env // Primitive value to NAPI call
	Global() (*Object, error)
	Undefined() (ValueType, error)
	Null() (ValueType, error)
}

// Return N-API env reference
func N_APIEnv(env napi.Env) EnvType { return &_Env{env} }

// N-API _Env
type _Env struct {
	NapiEnv napi.Env
}

// Return [napi.Env] to point from internal napi cgo
func (e *_Env) NapiValue() napi.Env {
	return e.NapiEnv
}

// Return representantion to 'This' [*Object]
func (e *_Env) Global() (*Object, error) {
	napiValue, err := mustValueErr(napi.GetGlobal(e.NapiEnv))
	if err != nil {
		return nil, err
	}
	return ToObject(N_APIValue(e, napiValue)), nil
}

// Return Undefined value
func (e *_Env) Undefined() (ValueType, error) {
	napiValue, err := mustValueErr(napi.GetUndefined(e.NapiEnv))
	if err != nil {
		return nil, err
	}
	return N_APIValue(e, napiValue), nil
}

// Return Null value
func (e *_Env) Null() (ValueType, error) {
	napiValue, err := mustValueErr(napi.GetNull(e.NapiEnv))
	if err != nil {
		return nil, err
	}
	return N_APIValue(e, napiValue), nil
}
