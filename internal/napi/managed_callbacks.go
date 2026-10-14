package napi

/*
#include "managed_callbacks.h"
*/
import "C"

import (
	"fmt"
	"runtime/cgo"
	"unsafe"
)

type managedCallbacks struct {
	method, getter, setter Callback
}

func newToken(value any) *C.napi_go_token {
	handle := cgo.NewHandle(value)
	token := C.napi_go_token_new(C.uintptr_t(handle))
	if token == nil {
		handle.Delete()
	}
	return token
}

func deleteToken(token *C.napi_go_token) {
	cgo.Handle(token.handle).Delete()
	C.napi_go_token_delete(token)
}

//export deleteManagedCallback
func deleteManagedCallback(handle C.uintptr_t) {
	cgo.Handle(handle).Delete()
}

//export executeManagedCallback
func executeManagedCallback(cEnv C.napi_env, info C.napi_callback_info, handle C.uintptr_t, kind C.int) C.napi_value {
	env := Env(cEnv)
	defer func() {
		if recovered := recover(); recovered != nil {
			ThrowError(env, "", fmt.Sprintf("napi callback panic: %v", recovered))
		}
	}()
	callbacks := cgo.Handle(handle).Value().(managedCallbacks)
	callback := callbacks.method
	switch kind {
	case 1:
		callback = callbacks.getter
	case 2:
		callback = callbacks.setter
	}
	return C.napi_value(callback(env, CallbackInfo(info)))
}

// Each ordinary function owns its token until GC. Class/property callbacks use
// environment cleanup hooks, so extracted methods and accessors remain callable
// even after their constructor/prototype is collected. No Go pointer is kept by C.
func createManagedFunction(env Env, name string, callback Callback) (Value, Status) {
	if callback == nil {
		return nil, StatusInvalidArg
	}
	token := newToken(managedCallbacks{method: callback})
	if token == nil {
		return nil, StatusGenericFailure
	}
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	var result C.napi_value
	status := Status(C.napi_create_function(C.napi_env(env), cname, C.size_t(len(name)),
		C.napi_callback(C.napi_go_method), unsafe.Pointer(token), &result))
	if status != StatusOK {
		deleteToken(token)
		return nil, status
	}
	status = Status(C.napi_add_finalizer(C.napi_env(env), result, unsafe.Pointer(token),
		C.node_api_basic_finalize(C.napi_go_callback_finalize), nil, nil))
	if status != StatusOK {
		// The function may already exist. Keep its callback valid until shutdown.
		if cleanup := Status(C.napi_add_env_cleanup_hook(C.napi_env(env),
			C.napi_cleanup_hook(C.napi_go_callback_cleanup), unsafe.Pointer(token))); cleanup != StatusOK {
			// A failed finalizer/cleanup registration must never expose the function.
			deleteToken(token)
		}
		return nil, status
	}
	return Value(result), StatusOK
}

type propertyDescriptors struct {
	ptr    *C.napi_property_descriptor
	names  []*C.char
	tokens []*C.napi_go_token
}

func (d *propertyDescriptors) free() {
	for _, name := range d.names {
		C.free(unsafe.Pointer(name))
	}
	C.free(unsafe.Pointer(d.ptr))
}

func (d *propertyDescriptors) rollback(env Env) {
	for _, token := range d.tokens {
		C.napi_remove_env_cleanup_hook(C.napi_env(env), C.napi_cleanup_hook(C.napi_go_callback_cleanup), unsafe.Pointer(token))
		deleteToken(token)
	}
	d.tokens = nil
}

func (d *propertyDescriptors) register(env Env, callbacks managedCallbacks) (*C.napi_go_token, Status) {
	token := newToken(callbacks)
	if token == nil {
		return nil, StatusGenericFailure
	}
	status := Status(C.napi_add_env_cleanup_hook(C.napi_env(env),
		C.napi_cleanup_hook(C.napi_go_callback_cleanup), unsafe.Pointer(token)))
	if status != StatusOK {
		deleteToken(token)
		return nil, status
	}
	d.tokens = append(d.tokens, token)
	return token, StatusOK
}

func (d *propertyDescriptors) marshal(env Env, properties []PropertyDescriptor) Status {
	if len(properties) == 0 {
		return StatusOK
	}
	d.ptr = (*C.napi_property_descriptor)(C.calloc(C.size_t(len(properties)), C.size_t(C.sizeof_napi_property_descriptor)))
	if d.ptr == nil {
		return StatusGenericFailure
	}
	descriptors := unsafe.Slice(d.ptr, len(properties))
	for i, property := range properties {
		p := &descriptors[i]
		if property.Name != nil {
			p.name = C.napi_value(property.Name)
		} else {
			name := C.CString(property.Utf8name)
			d.names = append(d.names, name)
			p.utf8name = name
		}
		p.value = C.napi_value(property.Value)
		p.attributes = C.napi_property_attributes(property.Attributes)
		if property.Method != nil || property.Getter != nil || property.Setter != nil {
			token, status := d.register(env, managedCallbacks{property.Method, property.Getter, property.Setter})
			if status != StatusOK {
				return status
			}
			p.data = unsafe.Pointer(token)
			if property.Method != nil {
				p.method = C.napi_callback(C.napi_go_method)
			}
			if property.Getter != nil {
				p.getter = C.napi_callback(C.napi_go_getter)
			}
			if property.Setter != nil {
				p.setter = C.napi_callback(C.napi_go_setter)
			}
		}
	}
	return StatusOK
}

func DefineClass(env Env, name string, constructor Callback, properties []PropertyDescriptor) (Value, Status) {
	if constructor == nil {
		return nil, StatusInvalidArg
	}
	descriptors := &propertyDescriptors{}
	defer descriptors.free()
	if status := descriptors.marshal(env, properties); status != StatusOK {
		descriptors.rollback(env)
		return nil, status
	}
	token, status := descriptors.register(env, managedCallbacks{method: constructor})
	if status != StatusOK {
		descriptors.rollback(env)
		return nil, status
	}
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	var result C.napi_value
	status = Status(C.napi_define_class(C.napi_env(env), cname, C.size_t(len(name)),
		C.napi_callback(C.napi_go_method), unsafe.Pointer(token), C.size_t(len(properties)), descriptors.ptr, &result))
	if status != StatusOK {
		descriptors.rollback(env)
		return nil, status
	}
	return Value(result), StatusOK
}

func defineManagedProperties(env Env, object Value, properties []PropertyDescriptor) Status {
	descriptors := &propertyDescriptors{}
	defer descriptors.free()
	if status := descriptors.marshal(env, properties); status != StatusOK {
		descriptors.rollback(env)
		return status
	}
	status := Status(C.napi_define_properties(C.napi_env(env), C.napi_value(object),
		C.size_t(len(properties)), descriptors.ptr))
	// Defining multiple properties may partially succeed. Keep callback tokens
	// until environment teardown even on failure, rather than leave dangling data.
	return status
}

type WrapFinalize func(data any)

type managedWrap struct {
	data     any
	finalize WrapFinalize
}

//export finalizeManagedWrap
func finalizeManagedWrap(handle C.uintptr_t) {
	h := cgo.Handle(handle)
	defer h.Delete()
	// A GC finalizer must not call APIs that execute JavaScript or let Go panics
	// cross the C boundary. User cleanup is deliberately independent of napi_env.
	defer func() { _ = recover() }()
	wrapped := h.Value().(managedWrap)
	if wrapped.finalize != nil {
		wrapped.finalize(wrapped.data)
	}
}

func WrapGoValue(env Env, object Value, data any, finalize WrapFinalize) Status {
	token := newToken(managedWrap{data, finalize})
	if token == nil {
		return StatusGenericFailure
	}
	status := Status(C.napi_go_wrap(C.napi_env(env), C.napi_value(object), token))
	if status != StatusOK {
		deleteToken(token)
	}
	return status
}

func UnwrapGoValue(env Env, object Value, remove bool) (any, Status) {
	var token *C.napi_go_token
	var removeFlag C.int
	if remove {
		removeFlag = 1
	}
	status := Status(C.napi_go_unwrap(C.napi_env(env), C.napi_value(object), &token, removeFlag))
	if status != StatusOK {
		return nil, status
	}
	wrapped := cgo.Handle(token.handle).Value().(managedWrap)
	if remove {
		deleteToken(token)
	}
	return wrapped.data, StatusOK
}
