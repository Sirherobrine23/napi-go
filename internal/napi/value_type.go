package napi

// #include <node/node_api.h>
import "C"

// Describes the type of a napi_value. This generally corresponds to the types described in Section 6.1 of the ECMAScript Language Specification.
// In addition to types in that section, napi_valuetype can also represent Functions and Objects with external data.
//
// A JavaScript value of type napi_external appears in JavaScript as a plain object such that no properties can be set on it, and no prototype.
type ValueType C.napi_valuetype

const (
	ValueTypeUndefined ValueType = C.napi_undefined
	ValueTypeNull      ValueType = C.napi_null
	ValueTypeBoolean   ValueType = C.napi_boolean
	ValueTypeNumber    ValueType = C.napi_number
	ValueTypeString    ValueType = C.napi_string
	ValueTypeSymbol    ValueType = C.napi_symbol
	ValueTypeObject    ValueType = C.napi_object
	ValueTypeFunction  ValueType = C.napi_function
	ValueTypeExternal  ValueType = C.napi_external
	ValueTypeBigint    ValueType = C.napi_bigint
)

func (v ValueType) String() string {
	switch v {
	case ValueTypeUndefined:
		return "undefined"
	case ValueTypeNull:
		return "null"
	case ValueTypeBoolean:
		return "boolean"
	case ValueTypeNumber:
		return "number"
	case ValueTypeString:
		return "string"
	case ValueTypeSymbol:
		return "symbol"
	case ValueTypeObject:
		return "object"
	case ValueTypeFunction:
		return "function"
	case ValueTypeExternal:
		return "external"
	case ValueTypeBigint:
		return "bigint"
	default:
		return "undefined"
	}
}

// This represents the underlying binary scalar datatype of the TypedArray.
// Elements of this enum correspond to Section 22.2 of the ECMAScript Language Specification.
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

func (vTyped TypedArrayType) String() string {
	switch vTyped {
	case TypedArrayInt8Array:
		return "Int8 Array"
	case TypedArrayUint8Array:
		return "Uint8 Array"
	case TypedArrayUint8ClampedArray:
		return "Int8 Array clamped"
	case TypedArrayInt16Array:
		return "Int16 Array"
	case TypedArrayUint16Array:
		return "Uint16 Array"
	case TypedArrayInt32Array:
		return "Int32 Array"
	case TypedArrayUint32Array:
		return "Uint32 Array"
	case TypedArrayFloat32Array:
		return "Float32 Array"
	case TypedArrayFloat64Array:
		return "Float64 Array"
	case TypedArrayBigInt64Array:
		return "Bigint64 Array"
	case TypedArrayBigUint64Array:
		return "Biguint64 Array"
	default:
		return "Unknown"
	}
}

func (vTyped TypedArrayType) Size() int {
	switch vTyped {
	case TypedArrayInt8Array, TypedArrayUint8Array, TypedArrayUint8ClampedArray:
		return 1
	case TypedArrayInt16Array, TypedArrayUint16Array:
		return 2
	case TypedArrayInt32Array, TypedArrayUint32Array, TypedArrayFloat32Array:
		return 4
	case TypedArrayFloat64Array, TypedArrayBigInt64Array, TypedArrayBigUint64Array:
		return 8
	default:
		return 0
	}
}
