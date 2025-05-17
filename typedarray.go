package napi

import (
	"unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type TypedArrayType = napi.TypedArrayType

const (
	TypedArrayInt8Array         TypedArrayType = napi.TypedArrayInt8Array
	TypedArrayUint8Array        TypedArrayType = napi.TypedArrayUint8Array
	TypedArrayUint8ClampedArray TypedArrayType = napi.TypedArrayUint8ClampedArray
	TypedArrayInt16Array        TypedArrayType = napi.TypedArrayInt16Array
	TypedArrayUint16Array       TypedArrayType = napi.TypedArrayUint16Array
	TypedArrayInt32Array        TypedArrayType = napi.TypedArrayInt32Array
	TypedArrayUint32Array       TypedArrayType = napi.TypedArrayUint32Array
	TypedArrayFloat32Array      TypedArrayType = napi.TypedArrayFloat32Array
	TypedArrayFloat64Array      TypedArrayType = napi.TypedArrayFloat64Array
	TypedArrayBigInt64Array     TypedArrayType = napi.TypedArrayBigInt64Array
	TypedArrayBigUint64Array    TypedArrayType = napi.TypedArrayBigUint64Array
)

type TypedArray struct {
	value
	typeArray TypedArrayType
}

// CreateTypedArray creates a new typed array of the specified type and length,
// backed by the provided ArrayBuffer. The function takes the environment (env),
// the typed array type (Type), the desired length, the byte offset within the
// buffer, and a pointer to the ArrayBuffer. It returns a pointer to the created
// TypedArray of type T, or an error if the operation fails.
//
// Parameters:
//   - env: The environment in which to create the typed array.
//   - Type: The type of the typed array to create (e.g., Int8Array, Float32Array).
//   - length: The number of elements in the typed array.
//   - byteOffset: The offset in bytes from the start of the ArrayBuffer.
//   - arrayValue: Pointer to the ArrayBuffer to use as the backing store.
func CreateTypedArray(env EnvType, Type TypedArrayType, length, byteOffset int, arrayValue *ArrayBuffer) (*TypedArray, error) {
	size, err := arrayValue.ByteLength()
	if err != nil {
		return nil, err
	}

	value, status := napi.CreateTypedArray(env.NapiValue(),
		napi.TypedArrayType(Type), size,
		arrayValue.NapiValue(), byteOffset,
	)

	if err = status.ToError(); err != nil {
		return nil, err
	}

	return &TypedArray{
		value:     N_APIValue(env, value),
		typeArray: Type,
	}, nil
}

// Type returns the type of the TypedArray as a TypedArrayType.
// This method is used to retrieve the specific type of the TypedArray instance.
// It returns the type as a TypedArrayType constant, which can be used to identify
// the kind of typed array (e.g., Int8Array, Float32Array, etc.).
// The function is a method of the TypedArray struct and does not take any parameters.
// It is a simple getter method that returns the type of the TypedArray.
func (typed TypedArray) Type() TypedArrayType { return TypedArrayType(typed.typeArray) }

// Get retrieves the underlying byte slice and associated ArrayBuffer from the TypedArray.
// It returns the data as a byte slice, a pointer to the ArrayBuffer, and an error if the operation fails.
// The function extracts the necessary information using napi.GetTypedArrayInfo and handles any status errors accordingly.
func (typed TypedArray) Get() (data []byte, arr *ArrayBuffer, err error) {
	// TypedArrayType, int, *byte, Value, int, Status
	_, _, dataPoint, value, byteOffset, status := napi.GetTypedArrayInfo(typed.NapiEnv(), typed.NapiValue())
	if err = status.ToError(); err != nil {
		return
	}
	data = unsafe.Slice(dataPoint, byteOffset)
	arr = ToArrayBuffer(N_APIValue(typed.Env(), value))
	return
}

// Return Generic Value to TypedArray
func ToTypedArray(value ValueType) *TypedArray {
	// TypedArrayType, int, *byte, Value, int, Status
	typeN, _, _, _, _, status := napi.GetTypedArrayInfo(value.NapiEnv(), value.NapiValue())
	if err := status.ToError(); err != nil {
		panic(err)
	}
	return &TypedArray{value: value, typeArray: typeN}
}
