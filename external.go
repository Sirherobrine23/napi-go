package napi

import (
	"unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type External struct{ value }

func ToExternal(o ValueType) *External { return &External{o} }

// CreateExternal creates a new JavaScript external value in the given N-API environment.
// It associates the provided data pointer with the external, and allows specifying a finalize
// callback and a finalize hint, which will be called when the external is garbage collected.
func CreateExternal(env EnvType, data unsafe.Pointer, finalize napi.Finalize, finalizeHint unsafe.Pointer) (*External, error) {
	napiValue, status := napi.CreateExternal(env.NapiValue(), data, finalize, finalizeHint)
	if err := status.ToError(); err != nil {
		return nil, err
	}
	return ToExternal(N_APIValue(env, napiValue)), nil
}

// Value retrieves the underlying unsafe.Pointer associated with the External value.
// Returns the pointer and an error, if any occurred during retrieval.
func (ext *External) Value() (unsafe.Pointer, error) {
	ptr, status := napi.GetValueExternal(ext.NapiEnv(), ext.NapiValue())
	if err := status.ToError(); err != nil {
		return nil, err
	}
	return ptr, nil
}
