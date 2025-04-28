package napi

import (
	"unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

// ArrayBuffer represents a JavaScript ArrayBuffer.
type ArrayBuffer struct{ value }

// ToArrayBuffer converts a ValueType to *ArrayBuffer.
// It assumes the underlying N-API value is an ArrayBuffer.
// Use IsArrayBuffer to check before converting.
func ToArrayBuffer(o ValueType) *ArrayBuffer { return &ArrayBuffer{o} }

// CreateArrayBuffer creates a new JavaScript ArrayBuffer instance.
// It also returns a pointer to the underlying byte buffer.
func CreateArrayBuffer(env EnvType, length int) (*ArrayBuffer, []byte, error) {
	napiValue, dataPtr, status := napi.CreateArrayBuffer(env.NapiValue(), length)
	if err := status.ToError(); err != nil {
		return nil, nil, err
	}
	var dataSlice []byte
	if dataPtr != nil && length > 0 {
		// Create a Go slice backed by the C memory.
		// Be cautious with this slice, as its lifetime is tied to the ArrayBuffer.
		dataSlice = unsafe.Slice(dataPtr, length)
	}
	return ToArrayBuffer(N_APIValue(env, napiValue)), dataSlice, nil
}

// CreateExternalArrayBuffer creates a JavaScript ArrayBuffer instance over an existing external data buffer.
// The caller is responsible for managing the lifetime of the data buffer.
// The finalize callback will be invoked when the ArrayBuffer is garbage collected.
func CreateExternalArrayBuffer(env EnvType, data []byte, finalize napi.Finalize, finalizeHint unsafe.Pointer) (*ArrayBuffer, error) {
	var dataPtr unsafe.Pointer
	if len(data) > 0 {
		dataPtr = unsafe.Pointer(&data[0])
	}
	napiValue, status := napi.CreateExternalArrayBuffer(env.NapiValue(), dataPtr, len(data), finalize, finalizeHint)
	if err := status.ToError(); err != nil {
		return nil, err
	}
	return ToArrayBuffer(N_APIValue(env, napiValue)), nil
}

// Info retrieves information about the ArrayBuffer, including its underlying data buffer and length.
func (ab *ArrayBuffer) Info() ([]byte, int, error) {
	dataPtr, length, status := napi.GetArrayBufferInfo(ab.NapiEnv(), ab.NapiValue())
	if err := status.ToError(); err != nil {
		return nil, 0, err
	}
	var dataSlice []byte
	if dataPtr != nil && length > 0 {
		// Create a Go slice backed by the C memory.
		dataSlice = unsafe.Slice(dataPtr, length)
	}
	return dataSlice, length, nil
}

// Data retrieves the underlying byte data of the ArrayBuffer as a Go slice.
// This is a convenience wrapper around Info().
func (ab *ArrayBuffer) Data() ([]byte, error) {
	data, _, err := ab.Info()
	return data, err
}

// ByteLength retrieves the length (in bytes) of the ArrayBuffer.
// This is a convenience wrapper around Info().
func (ab *ArrayBuffer) ByteLength() (int, error) {
	_, length, err := ab.Info()
	return length, err
}

// Detach detaches the ArrayBuffer, making its contents inaccessible from JavaScript.
// This is used for transferring ownership of the underlying buffer.
func (ab *ArrayBuffer) Detach() error {
	return singleMustValueErr(napi.DetachArrayBuffer(ab.NapiEnv(), ab.NapiValue()))
}

// IsDetached checks if the ArrayBuffer has been detached.
func (ab *ArrayBuffer) IsDetached() (bool, error) {
	return mustValueErr(napi.IsDetachedArrayBuffer(ab.NapiEnv(), ab.NapiValue()))
}
