package napi

import (
	"unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

// DataView represents a JavaScript DataView object.
type DataView struct{ value }

// ToDataView converts a ValueType to *DataView.
// It assumes the underlying N-API value is a DataView.
// Use IsDataView to check before converting.
func ToDataView(o ValueType) *DataView { return &DataView{o} }

// CreateDataView creates a new JavaScript DataView instance over an existing ArrayBuffer.
func CreateDataView(env EnvType, buffer *ArrayBuffer, byteOffset, byteLength int) (*DataView, error) {
	napiValue, status := napi.CreateDataView(env.NapiValue(), byteLength, buffer.NapiValue(), byteOffset)
	if err := status.ToError(); err != nil {
		return nil, err
	}
	return ToDataView(N_APIValue(env, napiValue)), nil
}

// Info retrieves information about the DataView, including the underlying ArrayBuffer,
// its byte length, and the byte offset within the buffer. It also returns a pointer
// to the start of the DataView's section within the buffer's data.
func (dv *DataView) Info() (buffer *ArrayBuffer, byteLength int, byteOffset int, dataPtr unsafe.Pointer, err error) {
	var napiBuffer napi.Value
	var lengthC int      // Use int directly as C.size_t conversion is handled internally
	var offsetC int      // Use int directly
	var dataRawPtr *byte // Pointer to the start of the *ArrayBuffer*'s data

	lengthC, dataRawPtr, napiBuffer, offsetC, status := napi.GetDataViewInfo(dv.NapiEnv(), dv.NapiValue())
	if err = status.ToError(); err != nil {
		return
	}

	byteLength = lengthC
	byteOffset = offsetC
	buffer = ToArrayBuffer(N_APIValue(dv.Env(), napiBuffer))

	// Calculate the pointer to the start of the DataView's section
	if dataRawPtr != nil {
		dataPtr = unsafe.Pointer(uintptr(unsafe.Pointer(dataRawPtr)) + uintptr(byteOffset))
	}

	return
}

// ByteLength returns the length (in bytes) of the DataView.
// This is a convenience wrapper around Info().
func (dv *DataView) ByteLength() (int, error) {
	_, length, _, _, err := dv.Info()
	return length, err
}

// ByteOffset returns the offset (in bytes) of the DataView within its underlying ArrayBuffer.
// This is a convenience wrapper around Info().
func (dv *DataView) ByteOffset() (int, error) {
	_, _, offset, _, err := dv.Info()
	return offset, err
}

// Buffer returns the underlying ArrayBuffer referenced by the DataView.
// This is a convenience wrapper around Info().
func (dv *DataView) Buffer() (*ArrayBuffer, error) {
	buffer, _, _, _, err := dv.Info()
	return buffer, err
}
