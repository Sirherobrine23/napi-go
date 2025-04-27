package napi

/*
#include <node/node_api.h>
*/
import "C"

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
