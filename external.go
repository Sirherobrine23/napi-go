package napi

import (
	"unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type External struct{ value }

func ToExternal(o ValueType) *External { return &External{o} }

func CreateExternal(env EnvType, data unsafe.Pointer, finalize napi.Finalize, finalizeHint unsafe.Pointer) (*External, error) {
	napiValue, status := napi.CreateExternal(env.NapiValue(), data, finalize, finalizeHint)
	if err := status.ToError(); err != nil {
		return nil, err
	}
	return ToExternal(N_APIValue(env, napiValue)), nil
}

func (ext *External) Value() (unsafe.Pointer, error) {
	ptr, status := napi.GetValueExternal(ext.NapiEnv(), ext.NapiValue())
	if err := status.ToError(); err != nil {
		return nil, err
	}
	return ptr, nil
}
