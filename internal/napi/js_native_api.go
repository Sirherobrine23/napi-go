package napi

// #include <stdlib.h>
// #include <node/node_api.h>
import "C"

import "unsafe"

type CallbackInfo unsafe.Pointer
type Callback func(env Env, info CallbackInfo) Value

func GetUndefined(env Env) (Value, Status) {
	var result Value
	status := Status(C.napi_get_undefined(
		C.napi_env(env),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetNull(env Env) (Value, Status) {
	var result Value
	status := Status(C.napi_get_null(
		C.napi_env(env),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetGlobal(env Env) (Value, Status) {
	var result Value
	status := Status(C.napi_get_global(
		C.napi_env(env),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetBoolean(env Env, value bool) (Value, Status) {
	var result Value
	status := Status(C.napi_get_boolean(
		C.napi_env(env),
		C.bool(value),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateObject(env Env) (Value, Status) {
	var result Value
	status := Status(C.napi_create_object(
		C.napi_env(env),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateArray(env Env) (Value, Status) {
	var result Value
	status := Status(C.napi_create_array(
		C.napi_env(env),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateArrayWithLength(env Env, length int) (Value, Status) {
	var result Value
	status := Status(C.napi_create_array_with_length(
		C.napi_env(env),
		C.size_t(length),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateDouble(env Env, value float64) (Value, Status) {
	var result Value
	status := Status(C.napi_create_double(
		C.napi_env(env),
		C.double(value),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateStringUtf8(env Env, str string) (Value, Status) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))

	var result Value
	status := Status(C.napi_create_string_utf8(
		C.napi_env(env),
		cstr,
		C.size_t(len([]byte(str))), // must pass number of bytes
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateSymbol(env Env, description Value) (Value, Status) {
	var result Value
	status := Status(C.napi_create_symbol(
		C.napi_env(env),
		C.napi_value(description),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateFunction(env Env, name string, cb Callback) (Value, Status) {
	provider, status := getInstanceData(env)
	if status != StatusOK || provider == nil {
		return nil, status
	}

	return provider.GetCallbackData().CreateCallback(env, name, cb)
}

func CreateError(env Env, code, msg Value) (Value, Status) {
	var result Value
	status := Status(C.napi_create_error(
		C.napi_env(env),
		C.napi_value(code),
		C.napi_value(msg),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func Typeof(env Env, value Value) (ValueType, Status) {
	var result ValueType
	status := Status(C.napi_typeof(
		C.napi_env(env),
		C.napi_value(value),
		(*C.napi_valuetype)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetValueDouble(env Env, value Value) (float64, Status) {
	var result float64
	status := Status(C.napi_get_value_double(
		C.napi_env(env),
		C.napi_value(value),
		(*C.double)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetValueBool(env Env, value Value) (bool, Status) {
	var result bool
	status := Status(C.napi_get_value_bool(
		C.napi_env(env),
		C.napi_value(value),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetValueStringUtf8(env Env, value Value) (string, Status) {
	// call napi_get_value_string_utf8 twice
	// first is to get number of bytes
	// second is to populate the actual string buffer
	bufsize := C.size_t(0)
	var strsize C.size_t

	status := Status(C.napi_get_value_string_utf8(
		C.napi_env(env),
		C.napi_value(value),
		nil,
		bufsize,
		&strsize,
	))

	if status != StatusOK {
		return "", status
	}

	// ensure there is room for the null terminator as well
	strsize++
	cstr := (*C.char)(C.malloc(C.sizeof_char * strsize))
	defer C.free(unsafe.Pointer(cstr))

	status = Status(C.napi_get_value_string_utf8(
		C.napi_env(env),
		C.napi_value(value),
		cstr,
		strsize,
		&strsize,
	))

	if status != StatusOK {
		return "", status
	}

	return C.GoStringN(
		(*C.char)(cstr),
		(C.int)(strsize),
	), status
}

func GetValueStringUtf16(env Env, value Value) ([]uint16, Status) {
	bufsize := C.size_t(0)
	var strsize C.size_t

	status := Status(C.napi_get_value_string_utf16(C.napi_env(env), C.napi_value(value), nil, bufsize, &strsize))
	if status != StatusOK {
		return nil, status
	}

	strsize++
	cstr := (*C.char16_t)(C.malloc(C.sizeof_char * strsize))
	defer C.free(unsafe.Pointer(cstr))

	status = Status(C.napi_get_value_string_utf16(C.napi_env(env), C.napi_value(value), cstr, strsize, &strsize))
	if status != StatusOK {
		return nil, status
	}

	return unsafe.Slice((*uint16)(cstr), strsize), status
}

func SetProperty(env Env, object, key, value Value) Status {
	return Status(C.napi_set_property(
		C.napi_env(env),
		C.napi_value(object),
		C.napi_value(key),
		C.napi_value(value),
	))
}

func SetElement(env Env, object Value, index int, value Value) Status {
	return Status(C.napi_set_element(
		C.napi_env(env),
		C.napi_value(object),
		C.uint32_t(index),
		C.napi_value(value),
	))
}

func StrictEquals(env Env, lhs, rhs Value) (bool, Status) {
	var result bool
	status := Status(C.napi_strict_equals(
		C.napi_env(env),
		C.napi_value(lhs),
		C.napi_value(rhs),
		(*C.bool)(&result),
	))
	return result, status
}

type GetCbInfoResult struct {
	Args []Value
	This Value
}

func GetCbInfo(env Env, info CallbackInfo) (GetCbInfoResult, Status) {
	// call napi_get_cb_info twice
	// first is to get total number of arguments
	// second is to populate the actual arguments
	argc := C.size_t(0)
	status := Status(C.napi_get_cb_info(
		C.napi_env(env),
		C.napi_callback_info(info),
		&argc,
		nil,
		nil,
		nil,
	))

	if status != StatusOK {
		return GetCbInfoResult{}, status
	}

	argv := make([]Value, int(argc))
	var cArgv unsafe.Pointer
	if argc > 0 {
		cArgv = unsafe.Pointer(&argv[0]) // must pass element pointer
	}

	var thisArg Value

	status = Status(C.napi_get_cb_info(
		C.napi_env(env),
		C.napi_callback_info(info),
		&argc,
		(*C.napi_value)(cArgv),
		(*C.napi_value)(unsafe.Pointer(&thisArg)),
		nil,
	))

	return GetCbInfoResult{
		Args: argv,
		This: thisArg,
	}, status
}

func Throw(env Env, err Value) Status {
	return Status(C.napi_throw(
		C.napi_env(env),
		C.napi_value(err),
	))
}

func ThrowError(env Env, code, msg string) Status {
	codeCStr, msgCCstr := C.CString(code), C.CString(msg)
	defer C.free(unsafe.Pointer(codeCStr))
	defer C.free(unsafe.Pointer(msgCCstr))

	return Status(C.napi_throw_error(
		C.napi_env(env),
		codeCStr,
		msgCCstr,
	))
}

func SetInstanceData(env Env, data any) Status {
	provider, status := getInstanceData(env)
	if status != StatusOK || provider == nil {
		return status
	}

	provider.SetUserData(data)
	return status
}

func GetInstanceData(env Env) (any, Status) {
	provider, status := getInstanceData(env)
	if status != StatusOK || provider == nil {
		return nil, status
	}

	return provider.GetUserData(), status
}

func CreateExternal(env Env, data unsafe.Pointer, finalize Finalize, finalizeHint unsafe.Pointer) (Value, Status) {
	var result Value
	finalizer := FinalizeToFinalizer(finalize)
	status := Status(C.napi_create_external(
		C.napi_env(env),
		data,
		C.napi_finalize(unsafe.Pointer(&finalizer)),
		finalizeHint,
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetValueInt32(env Env, value Value) (int32, Status) {
	var result C.int32_t
	status := Status(C.napi_get_value_int32(
		C.napi_env(env),
		C.napi_value(value),
		&result,
	))
	return int32(result), status
}

func GetValueUint32(env Env, value Value) (uint32, Status) {
	var result uint32
	status := Status(C.napi_get_value_uint32(
		C.napi_env(env),
		C.napi_value(value),
		(*C.uint32_t)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetValueInt64(env Env, value Value) (int64, Status) {
	var result int64
	status := Status(C.napi_get_value_int64(
		C.napi_env(env),
		C.napi_value(value),
		(*C.int64_t)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetValueBigIntInt64(env Env, value Value) (int64, bool, Status) {
	var result int64
	var lossless bool
	status := Status(C.napi_get_value_bigint_int64(
		C.napi_env(env),
		C.napi_value(value),
		(*C.int64_t)(unsafe.Pointer(&result)),
		(*C.bool)(unsafe.Pointer(&lossless)),
	))
	return result, lossless, status
}

func GetValueBigIntWords(env Env, value Value, signBit int, wordCount int, words *uint64) Status {
	return Status(C.napi_get_value_bigint_words(
		C.napi_env(env),
		C.napi_value(value),
		(*C.int)(unsafe.Pointer(&signBit)),
		(*C.size_t)(unsafe.Pointer(&wordCount)),
		(*C.uint64_t)(unsafe.Pointer(words)),
	))
}

func GetValueExternal(env Env, value Value) (unsafe.Pointer, Status) {
	var result unsafe.Pointer
	status := Status(C.napi_get_value_external(
		C.napi_env(env),
		C.napi_value(value),
		&result,
	))
	return result, status
}

func CoerceToBool(env Env, value Value) (Value, Status) {
	var result Value
	status := Status(C.napi_coerce_to_bool(
		C.napi_env(env),
		C.napi_value(value),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CoerceToNumber(env Env, value Value) (Value, Status) {
	var result Value
	status := Status(C.napi_coerce_to_number(
		C.napi_env(env),
		C.napi_value(value),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CoerceToObject(env Env, value Value) (Value, Status) {
	var result Value
	status := Status(C.napi_coerce_to_object(
		C.napi_env(env),
		C.napi_value(value),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CoerceToString(env Env, value Value) (Value, Status) {
	var result Value
	status := Status(C.napi_coerce_to_string(
		C.napi_env(env),
		C.napi_value(value),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateBuffer(env Env, length int) (Value, Status) {
	var result Value
	status := Status(C.napi_create_buffer(
		C.napi_env(env),
		C.size_t(length),
		nil,
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateBufferCopy(env Env, data []byte) (Value, Status) {
	var result Value
	status := Status(C.napi_create_buffer_copy(
		C.napi_env(env),
		C.size_t(len(data)),
		unsafe.Pointer(&data[0]),
		nil,
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetBufferInfo(env Env, value Value) (*byte, int, Status) {
	var data *byte
	var length C.size_t

	dataPtr := unsafe.Pointer(&data)
	status := Status(C.napi_get_buffer_info(C.napi_env(env), C.napi_value(value), &dataPtr, &length))
	return data, int(length), status
}

func GetBufferInfoSize(env Env, value Value) (int, Status) {
	var length C.size_t
	status := Status(C.napi_get_buffer_info(C.napi_env(env), C.napi_value(value), nil, &length))
	return int(length), status
}

func GetBufferInfoData(env Env, value Value) (buff []byte, status Status) {
	var data *byte
	var length C.size_t

	dataPtr := unsafe.Pointer(&data)
	status = Status(C.napi_get_buffer_info(C.napi_env(env), C.napi_value(value), &dataPtr, &length))
	if status == StatusOK {
		buff = unsafe.Slice(data, length)
	}
	return
}

func GetArrayLength(env Env, value Value) (int, Status) {
	var length C.uint32_t
	status := Status(C.napi_get_array_length(
		C.napi_env(env),
		C.napi_value(value),
		&length,
	))
	return int(length), status
}

func GetPrototype(env Env, value Value) (Value, Status) {
	var result Value
	status := Status(C.napi_get_prototype(
		C.napi_env(env),
		C.napi_value(value),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func InstanceOf(env Env, object, constructor Value) (bool, Status) {
	var result bool
	status := Status(C.napi_instanceof(
		C.napi_env(env),
		C.napi_value(object),
		C.napi_value(constructor),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func IsArray(env Env, value Value) (bool, Status) {
	var result bool
	status := Status(C.napi_is_array(
		C.napi_env(env),
		C.napi_value(value),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func IsBuffer(env Env, value Value) (bool, Status) {
	var result bool
	status := Status(C.napi_is_buffer(
		C.napi_env(env),
		C.napi_value(value),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func IsError(env Env, value Value) (bool, Status) {
	var result bool
	status := Status(C.napi_is_error(
		C.napi_env(env),
		C.napi_value(value),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func IsPromise(env Env, value Value) (bool, Status) {
	var result bool
	status := Status(C.napi_is_promise(
		C.napi_env(env),
		C.napi_value(value),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func IsTypedArray(env Env, value Value) (bool, Status) {
	var result bool
	status := Status(C.napi_is_typedarray(
		C.napi_env(env),
		C.napi_value(value),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetTypedArrayInfo(env Env, value Value) (TypedArrayType, int, *byte, Value, int, Status) {
	var type_ TypedArrayType
	var length C.size_t
	var data *byte
	var arrayBuffer Value
	var byteOffset C.size_t

	dataPtr := unsafe.Pointer(&data)
	status := Status(C.napi_get_typedarray_info(
		C.napi_env(env),
		C.napi_value(value),
		(*C.napi_typedarray_type)(unsafe.Pointer(&type_)),
		&length,
		&dataPtr,
		(*C.napi_value)(unsafe.Pointer(&arrayBuffer)),
		&byteOffset,
	))
	return type_, int(length), data, arrayBuffer, int(byteOffset), status
}

func CreateTypedArray(env Env, type_ TypedArrayType, length int, arrayBuffer Value, byteOffset int) (Value, Status) {
	var result Value
	status := Status(C.napi_create_typedarray(
		C.napi_env(env),
		C.napi_typedarray_type(type_),
		C.size_t(length),
		C.napi_value(arrayBuffer),
		C.size_t(byteOffset),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func AdjustExternalMemory(env Env, change int64) (int64, Status) {
	var result int64
	status := Status(C.napi_adjust_external_memory(
		C.napi_env(env),
		C.int64_t(change),
		(*C.int64_t)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateDataView(env Env, length int, arrayBuffer Value, byteOffset int) (Value, Status) {
	var result Value
	status := Status(C.napi_create_dataview(
		C.napi_env(env),
		C.size_t(length),
		C.napi_value(arrayBuffer),
		C.size_t(byteOffset),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetDataViewInfo(env Env, value Value) (int, *byte, Value, int, Status) {
	var length C.size_t
	var data *byte
	var arrayBuffer Value
	var byteOffset C.size_t

	dataPtr := unsafe.Pointer(&data)
	status := Status(C.napi_get_dataview_info(
		C.napi_env(env),
		C.napi_value(value),
		&length,
		&dataPtr,
		(*C.napi_value)(unsafe.Pointer(&arrayBuffer)),
		&byteOffset,
	))
	return int(length), data, arrayBuffer, int(byteOffset), status
}

func GetAllPropertyNames(env Env, object Value, keyMode KeyCollectionMode, keyFilter KeyFilter, keyConversion KeyConversion) (Value, Status) {
	var result Value
	status := Status(C.napi_get_all_property_names(
		C.napi_env(env),
		C.napi_value(object),
		C.napi_key_collection_mode(keyMode),
		C.napi_key_filter(keyFilter),
		C.napi_key_conversion(keyConversion),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func HasOwnProperty(env Env, object, key Value) (bool, Status) {
	var result bool
	status := Status(C.napi_has_own_property(
		C.napi_env(env),
		C.napi_value(object),
		C.napi_value(key),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func HasProperty(env Env, object, key Value) (bool, Status) {
	var result bool
	status := Status(C.napi_has_property(
		C.napi_env(env),
		C.napi_value(object),
		C.napi_value(key),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetPropertyNames(env Env, object Value) (Value, Status) {
	var result Value
	status := Status(C.napi_get_property_names(
		C.napi_env(env),
		C.napi_value(object),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func DefineProperties(env Env, object Value, properties []PropertyDescriptor) Status {
	return Status(C.napi_define_properties(
		C.napi_env(env),
		C.napi_value(object),
		C.size_t(len(properties)),
		(*C.napi_property_descriptor)(unsafe.Pointer(&properties[0])),
	))
}

func GetValueBigIntUint64(env Env, value Value) (uint64, bool, Status) {
	var result uint64
	var lossless bool
	status := Status(C.napi_get_value_bigint_uint64(
		C.napi_env(env),
		C.napi_value(value),
		(*C.uint64_t)(unsafe.Pointer(&result)),
		(*C.bool)(unsafe.Pointer(&lossless)),
	))
	return result, lossless, status
}

func CreateBigIntInt64(env Env, value int64) (Value, Status) {
	var result Value
	status := Status(C.napi_create_bigint_int64(
		C.napi_env(env),
		C.int64_t(value),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateBigIntUint64(env Env, value uint64) (Value, Status) {
	var result Value
	status := Status(C.napi_create_bigint_uint64(
		C.napi_env(env),
		C.uint64_t(value),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateBigIntWords(env Env, signBit int, wordCount int, words *uint64) (Value, Status) {
	var result Value
	status := Status(C.napi_create_bigint_words(
		C.napi_env(env),
		C.int(signBit),
		C.size_t(wordCount),
		(*C.uint64_t)(unsafe.Pointer(words)),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func IsDate(env Env, value Value) (bool, Status) {
	var result bool
	status := Status(C.napi_is_date(
		C.napi_env(env),
		C.napi_value(value),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func IsDetachedArrayBuffer(env Env, value Value) (bool, Status) {
	var result bool
	status := Status(C.napi_is_detached_arraybuffer(
		C.napi_env(env),
		C.napi_value(value),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func DetachArrayBuffer(env Env, value Value) Status {
	return Status(C.napi_detach_arraybuffer(
		C.napi_env(env),
		C.napi_value(value),
	))
}

func CreateArrayBuffer(env Env, length int) (Value, *byte, Status) {
	var result Value
	var data *byte
	dataPtr := unsafe.Pointer(&data)
	status := Status(C.napi_create_arraybuffer(
		C.napi_env(env),
		C.size_t(length),
		&dataPtr,
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, data, status
}

func GetArrayBufferInfo(env Env, value Value) (*byte, int, Status) {
	var data *byte
	var length C.size_t
	dataPtr := unsafe.Pointer(&data)
	status := Status(C.napi_get_arraybuffer_info(
		C.napi_env(env),
		C.napi_value(value),
		&dataPtr,
		&length,
	))
	return data, int(length), status
}

func CreateExternalArrayBuffer(env Env, data unsafe.Pointer, length int, finalize Finalize, finalizeHint unsafe.Pointer) (Value, Status) {
	var result Value
	finalizer := FinalizeToFinalizer(finalize)
	status := Status(C.napi_create_external_arraybuffer(
		C.napi_env(env),
		data,
		C.size_t(length),
		C.napi_finalize(unsafe.Pointer(&finalizer)),
		finalizeHint,
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetElement(env Env, object Value, index int) (Value, Status) {
	var result Value
	status := Status(C.napi_get_element(
		C.napi_env(env),
		C.napi_value(object),
		C.uint32_t(index),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetProperty(env Env, object, key Value) (Value, Status) {
	var result Value
	status := Status(C.napi_get_property(
		C.napi_env(env),
		C.napi_value(object),
		C.napi_value(key),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func DeleteProperty(env Env, object, key Value) (bool, Status) {
	var result bool
	status := Status(C.napi_delete_property(
		C.napi_env(env),
		C.napi_value(object),
		C.napi_value(key),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func SetNamedProperty(env Env, object Value, name string, value Value) Status {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return Status(C.napi_set_named_property(
		C.napi_env(env),
		C.napi_value(object),
		cname,
		C.napi_value(value),
	))
}

func GetNamedProperty(env Env, object Value, name string) (Value, Status) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	var result Value
	status := Status(C.napi_get_named_property(
		C.napi_env(env),
		C.napi_value(object),
		cname,
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func HasNamedProperty(env Env, object Value, name string) (bool, Status) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	var result bool
	status := Status(C.napi_has_named_property(
		C.napi_env(env),
		C.napi_value(object),
		cname,
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func HasElement(env Env, object Value, index int) (bool, Status) {
	var result bool
	status := Status(C.napi_has_element(
		C.napi_env(env),
		C.napi_value(object),
		C.uint32_t(index),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func DeleteElement(env Env, object Value, index int) (bool, Status) {
	var result bool
	status := Status(C.napi_delete_element(
		C.napi_env(env),
		C.napi_value(object),
		C.uint32_t(index),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func ObjectFreeze(env Env, object Value) Status {
	return Status(C.napi_object_freeze(
		C.napi_env(env),
		C.napi_value(object),
	))
}

func ObjectSeal(env Env, object Value) Status {
	return Status(C.napi_object_seal(
		C.napi_env(env),
		C.napi_value(object),
	))
}

func ThrowTypeError(env Env, code, msg string) Status {
	codeCStr, msgCCstr := C.CString(code), C.CString(msg)
	defer C.free(unsafe.Pointer(codeCStr))
	defer C.free(unsafe.Pointer(msgCCstr))

	return Status(C.napi_throw_type_error(
		C.napi_env(env),
		codeCStr,
		msgCCstr,
	))
}

func ThrowRangeError(env Env, code, msg string) Status {
	codeCStr, msgCCstr := C.CString(code), C.CString(msg)
	defer C.free(unsafe.Pointer(codeCStr))
	defer C.free(unsafe.Pointer(msgCCstr))

	return Status(C.napi_throw_range_error(
		C.napi_env(env),
		codeCStr,
		msgCCstr,
	))
}

func CreateTypeError(env Env, code, msg Value) (Value, Status) {
	var result Value
	status := Status(C.napi_create_type_error(
		C.napi_env(env),
		C.napi_value(code),
		C.napi_value(msg),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateRangeError(env Env, code, msg Value) (Value, Status) {
	var result Value
	status := Status(C.napi_create_range_error(
		C.napi_env(env),
		C.napi_value(code),
		C.napi_value(msg),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func IsExceptionPending(env Env) (bool, Status) {
	var result bool
	status := Status(C.napi_is_exception_pending(
		C.napi_env(env),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetAndClearLastException(env Env) (Value, Status) {
	var result Value
	status := Status(C.napi_get_and_clear_last_exception(
		C.napi_env(env),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

type CallbackScope struct {
	scope C.napi_callback_scope
}

// CloseCallbackScope Function to close a callback scope
func CloseCallbackScope(env Env, scope CallbackScope) Status {
	return Status(C.napi_close_callback_scope(
		C.napi_env(env),
		scope.scope,
	))
}

func CreateInt32(env Env, value int32) (Value, Status) {
	var result Value
	status := Status(C.napi_create_int32(
		C.napi_env(env),
		C.int32_t(value),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateUint32(env Env, value uint32) (Value, Status) {
	var result Value
	status := Status(C.napi_create_uint32(
		C.napi_env(env),
		C.uint32_t(value),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateInt64(env Env, value int64) (Value, Status) {
	var result Value
	status := Status(C.napi_create_int64(
		C.napi_env(env),
		C.int64_t(value),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateStringLatin1(env Env, str string) (Value, Status) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))

	var result Value
	status := Status(C.napi_create_string_latin1(
		C.napi_env(env),
		cstr,
		C.size_t(len([]byte(str))),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateStringUtf16(env Env, str []uint16) (Value, Status) {
	var result Value
	status := Status(C.napi_create_string_utf16(
		C.napi_env(env),
		(*C.char16_t)(unsafe.Pointer(&str[0])),
		C.size_t(len(str)),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CallFunction(env Env, recv Value, fn Value, argc int, argv []Value) (Value, Status) {
	var result Value
	status := Status(C.napi_call_function(
		C.napi_env(env),
		C.napi_value(recv),
		C.napi_value(fn),
		C.size_t(argc),
		(*C.napi_value)(unsafe.Pointer(&argv[0])),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetNewTarget(env Env, info CallbackInfo) (Value, Status) {
	var result Value
	status := Status(C.napi_get_new_target(
		C.napi_env(env),
		C.napi_callback_info(info),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func NewInstance(env Env, constructor Value, argc int, argv []Value) (Value, Status) {
	var result Value
	status := Status(C.napi_new_instance(
		C.napi_env(env),
		C.napi_value(constructor),
		C.size_t(argc),
		(*C.napi_value)(unsafe.Pointer(&argv[0])),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func IsDataView(env Env, value Value) (bool, Status) {
	var result bool
	status := Status(C.napi_is_dataview(
		C.napi_env(env),
		C.napi_value(value),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func IsArrayBuffer(env Env, value Value) (bool, Status) {
	var result bool
	status := Status(C.napi_is_arraybuffer(
		C.napi_env(env),
		C.napi_value(value),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetDateValue(env Env, value Value) (float64, Status) {
	var result float64
	status := Status(C.napi_get_date_value(
		C.napi_env(env),
		C.napi_value(value),
		(*C.double)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateDate(env Env, time float64) (Value, Status) {
	var result Value
	status := Status(C.napi_create_date(
		C.napi_env(env),
		(C.double)(time),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}
