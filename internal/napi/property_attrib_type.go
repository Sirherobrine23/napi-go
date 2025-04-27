package napi

// #include <node/node_api.h>
import "C"
import "unsafe"

// napi_property_attributes are flags used to control the behavior of properties set on a JavaScript object.
// Other than napi_static they correspond to the attributes listed in Section 6.1.7.1 of the ECMAScript Language Specification.
type PropertyAttributes C.napi_property_attributes

const (
	Default           PropertyAttributes = C.napi_default
	Writable          PropertyAttributes = C.napi_writable
	Enumerable        PropertyAttributes = C.napi_enumerable
	Configurable      PropertyAttributes = C.napi_configurable
	Static            PropertyAttributes = C.napi_static
	DefaultMethod     PropertyAttributes = C.napi_default_method
	DefaultJSProperty PropertyAttributes = C.napi_default_jsproperty
)

type PropertyDescriptor struct {
	Utf8name   string
	Name       Value
	Method     Callback
	Getter     Callback
	Setter     Callback
	Value      Value
	Attributes PropertyAttributes
	Data       unsafe.Pointer
}
