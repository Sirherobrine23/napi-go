package napi

// wrapDecodedValue chooses a wrapper for a known JavaScript type. Assignment to
// a concrete Go wrapper still checks compatibility in valueFromDepth.
func wrapDecodedValue(value ValueType, kind NapiType) ValueType {
	switch kind {
	case TypeBoolean:
		return ToBoolean(value)
	case TypeNumber:
		return ToNumber(value)
	case TypeBigInt:
		return ToBigint(value)
	case TypeString:
		return ToString(value)
	case TypeObject:
		return ToObject(value)
	case TypeFunction:
		return ToFunction(value)
	case TypeExternal:
		return ToExternal(value)
	case TypeTypedArray:
		// TypedArray.Type currently reports its element kind instead of NapiType,
		// so the wrapper does not implement ValueType. Preserve the raw handle.
		return value
	case TypePromise:
		// A JavaScript-owned Promise has no napi_deferred. Keep its raw handle;
		// only CreatePromise can provide a resolvable *Promise wrapper.
		return value
	case TypeBuffer:
		return ToBuffer(value)
	case TypeDate:
		return ToDate(value)
	case TypeArray:
		return ToArray(value)
	case TypeArrayBuffer:
		return ToArrayBuffer(value)
	case TypeDataView:
		return ToDataView(value)
	case TypeError:
		return ToError(value)
	default:
		return value
	}
}
