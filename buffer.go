package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

type Buffer struct{ value }

// Convert [ValueType] to [*Buffer].
func ToBuffer(o ValueType) *Buffer { return &Buffer{o} }

// Create new Buffer with length
func CreateBuffer(env EnvType, length int) (*Buffer, error) {
	napiValue, err := mustValueErr(napi.CreateBuffer(env.NapiValue(), length))
	if err != nil {
		return nil, err
	}
	return ToBuffer(N_APIValue(env, napiValue)), nil
}

// Copy []byte to Node::Buffer struct
func CopyBuffer(env EnvType, buff []byte) (*Buffer, error) {
	napiValue, err := mustValueErr(napi.CreateBufferCopy(env.NapiValue(), buff))
	if err != nil {
		return nil, err
	}
	return ToBuffer(N_APIValue(env, napiValue)), nil
}

// Get size of buffer
func (buff *Buffer) Length() (int, error) {
	return mustValueErr(napi.GetBufferInfoSize(buff.NapiEnv(), buff.NapiValue()))
}

// return []byte from Buffer value
func (buff *Buffer) Data() ([]byte, error) {
	return mustValueErr(napi.GetBufferInfoData(buff.NapiEnv(), buff.NapiValue()))
}
