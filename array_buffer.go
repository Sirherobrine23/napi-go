package napi

import (
	"unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type ArrayBuffer struct{ *Value }

func CreateArrayBuffer(env *Env, buff []byte) (*ArrayBuffer, error) {
	value, point, err := napi.MustValueErr3(napi.CreateArrayBuffer(env.NapiValue(), len(buff)))
	if err != nil {
		return nil, err
	}

	// Copy data from the byte slice to the pointer
	copy((*[1 << 30]byte)(unsafe.Pointer(point))[:len(buff):len(buff)], buff)
	return &ArrayBuffer{
		Value: &Value{
			env:     env,
			valueOf: value,
			typeof:  napi.MustValue(napi.Typeof(env.NapiValue(), value)),
		},
	}, nil
}

func (buff *ArrayBuffer) ByteLenght() (int, error) {
	_, size, err := napi.MustValueErr3(napi.GetArrayBufferInfo(buff.NapiEnv(), buff.NapiValue()))
	if err != nil {
		return 0, err
	}
	return size, nil
}

func (buff *ArrayBuffer) Data() ([]byte, error) {
	bytePoint, size, err := napi.MustValueErr3(napi.GetArrayBufferInfo(buff.NapiEnv(), buff.NapiValue()))
	if err != nil {
		return nil, err
	}

	// Copy data from the pointer to a byte slice
	buffRead := make([]byte, size)
	copy(buffRead, (*[1 << 30]byte)(unsafe.Pointer(bytePoint))[:size:size])

	return buffRead, nil
}
