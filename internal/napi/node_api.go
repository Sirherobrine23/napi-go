package napi

// #include <stdlib.h>
// #include <node/node_api.h>
import "C"

import "unsafe"

type NodeVersion struct {
	Major   uint
	Minor   uint
	Patch   uint
	Release string
}

func GetNodeVersion(env Env) (NodeVersion, Status) {
	var cresult *C.napi_node_version
	status := Status(C.napi_get_node_version(
		C.napi_env(env),
		(**C.napi_node_version)(&cresult),
	))

	if status != StatusOK {
		return NodeVersion{}, status
	}

	return NodeVersion{
		Major:   uint(cresult.major),
		Minor:   uint(cresult.minor),
		Patch:   uint(cresult.patch),
		Release: C.GoString(cresult.release),
	}, status
}

func GetModuleFileName(env Env) (string, Status) {
	var cresult *C.char
	status := Status(C.node_api_get_module_file_name(
		C.napi_env(env),
		(**C.char)(&cresult),
	))

	if status != StatusOK {
		return "", status
	}

	return C.GoString(cresult), status
}

func ThrowSyntaxError(env Env, code, msg string) Status {
	codeCStr, msgCStr := C.CString(code), C.CString(msg)
	defer C.free(unsafe.Pointer(codeCStr))
	defer C.free(unsafe.Pointer(msgCStr))

	return Status(C.node_api_throw_syntax_error(
		C.napi_env(env),
		codeCStr,
		msgCStr,
	))
}

func CreateSyntaxError(env Env, code, msg Value) (Value, Status) {
	var result Value
	status := Status(C.node_api_create_syntax_error(
		C.napi_env(env),
		C.napi_value(code),
		C.napi_value(msg),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func SymbolFor(env Env, description string) (Value, Status) {
	var result Value
	status := Status(C.node_api_symbol_for(
		C.napi_env(env),
		C.CString(description),
		C.size_t(len(description)),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreatePropertyKeyLatin1(env Env, str string) (Value, Status) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))

	var result Value
	status := Status(C.node_api_create_property_key_latin1(
		C.napi_env(env),
		cstr,
		C.size_t(len([]byte(str))),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreatePropertyKeyUtf16(env Env, str []uint16) (Value, Status) {
	var result Value
	status := Status(C.node_api_create_property_key_utf16(
		C.napi_env(env),
		(*C.char16_t)(unsafe.Pointer(&str[0])),
		C.size_t(len(str)),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreatePropertyKeyUtf8(env Env, str string) (Value, Status) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))

	var result Value
	status := Status(C.node_api_create_property_key_utf8(
		C.napi_env(env),
		cstr,
		C.size_t(len([]byte(str))),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}
