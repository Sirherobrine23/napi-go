package napi

/*
#include <node/node_api.h>
*/
import "C"

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
