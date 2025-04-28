package napi

// #include <node/node_api.h>
import "C"

import "unsafe"

func DefineClass(env Env, name string, constructor Callback, Property []PropertyDescriptor) (Value, Status) {
	var pro *C.napi_property_descriptor
	if len(Property) > 0 {
		pro = (*C.napi_property_descriptor)(unsafe.Pointer(&Property[0]))
	}

	var result Value
	call := &constructor
	status := Status(C.napi_define_class(
		C.napi_env(env),
		C.CString(name),
		C.size_t(len(name)),
		C.napi_callback(unsafe.Pointer(call)),
		nil,
		C.size_t(len(Property)),
		pro,
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}
