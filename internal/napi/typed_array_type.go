package napi

/*
#include <node/node_api.h>
*/
import "C"

type TypedArrayType C.napi_typedarray_type

const (
	TypedArrayInt8Array         TypedArrayType = C.napi_int8_array
	TypedArrayUint8Array        TypedArrayType = C.napi_uint8_array
	TypedArrayUint8ClampedArray TypedArrayType = C.napi_uint8_clamped_array
	TypedArrayInt16Array        TypedArrayType = C.napi_int16_array
	TypedArrayUint16Array       TypedArrayType = C.napi_uint16_array
	TypedArrayInt32Array        TypedArrayType = C.napi_int32_array
	TypedArrayUint32Array       TypedArrayType = C.napi_uint32_array
	TypedArrayFloat32Array      TypedArrayType = C.napi_float32_array
	TypedArrayFloat64Array      TypedArrayType = C.napi_float64_array
	TypedArrayBigInt64Array     TypedArrayType = C.napi_bigint64_array
	TypedArrayBigUint64Array    TypedArrayType = C.napi_biguint64_array
)
