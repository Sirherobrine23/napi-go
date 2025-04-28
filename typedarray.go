package napi

import (
	"fmt"
	"unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type TypedArray struct{ value }

func ToTypedArray(o ValueType) *TypedArray { return &TypedArray{o} }

func CreateTypedArray(env EnvType, type_ napi.TypedArrayType, length int, arrayBuffer *ArrayBuffer, byteOffset int) (*TypedArray, error) {
	if arrayBuffer == nil {
		return nil, fmt.Errorf("arrayBuffer cannot be nil")
	}

	napiValue, status := napi.CreateTypedArray(env.NapiValue(), type_, length, arrayBuffer.NapiValue(), byteOffset)
	if err := status.ToError(); err != nil {
		return nil, err
	}
	return ToTypedArray(N_APIValue(env, napiValue)), nil
}

func (ta *TypedArray) Info() (type_ napi.TypedArrayType, length int, dataPtr unsafe.Pointer, buffer *ArrayBuffer, byteOffset int, err error) {
	var napiBuffer napi.Value
	var lengthC int
	var offsetC int
	var dataRawPtr *byte
	var typeC napi.TypedArrayType

	typeC, lengthC, dataRawPtr, napiBuffer, offsetC, status := napi.GetTypedArrayInfo(ta.NapiEnv(), ta.NapiValue())
	if err = status.ToError(); err != nil {
		return
	}

	type_ = typeC
	length = lengthC
	byteOffset = offsetC
	buffer = ToArrayBuffer(N_APIValue(ta.Env(), napiBuffer))

	if dataRawPtr != nil && buffer != nil {
		bufferDataPtr, _, infoErr := napi.GetArrayBufferInfo(buffer.NapiEnv(), buffer.NapiValue())
		if infoErr.ToError() != nil {
			err = infoErr.ToError()
			return
		}
		if bufferDataPtr != nil {
			dataPtr = unsafe.Pointer(uintptr(unsafe.Pointer(bufferDataPtr)) + uintptr(byteOffset))
		}
	}

	return
}

func (ta *TypedArray) Type() (napi.TypedArrayType, error) {
	type_, _, _, _, _, err := ta.Info()
	return type_, err
}

func (ta *TypedArray) Length() (int, error) {
	_, length, _, _, _, err := ta.Info()
	return length, err
}

func (ta *TypedArray) ByteLength() (int, error) {
	type_, length, _, _, _, err := ta.Info()
	if err != nil {
		return 0, err
	}
	elementSize := type_.Size()
	return length * elementSize, nil
}

func (ta *TypedArray) ByteOffset() (int, error) {
	_, _, _, _, offset, err := ta.Info()
	return offset, err
}

func (ta *TypedArray) Buffer() (*ArrayBuffer, error) {
	_, _, _, buffer, _, err := ta.Info()
	return buffer, err
}

func (ta *TypedArray) Data() ([]byte, error) {
	_, _, dataPtr, _, _, err := ta.Info()
	if err != nil {
		return nil, err
	}
	if dataPtr == nil {
		return nil, nil
	}

	byteLen, err := ta.ByteLength()
	if err != nil {
		return nil, err
	}
	return unsafe.Slice((*byte)(dataPtr), byteLen), nil
}
