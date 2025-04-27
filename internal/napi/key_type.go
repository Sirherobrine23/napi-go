package napi

/*
#include <node/node_api.h>
*/
import "C"

// KeyCollectionMode type
type KeyCollectionMode C.napi_key_collection_mode

const (
	KeyIncludePrototypes KeyCollectionMode = C.napi_key_include_prototypes
	KeyOwnOnly           KeyCollectionMode = C.napi_key_own_only
)

// KeyFilter type
type KeyFilter C.napi_key_filter

const (
	KeyAllProperties KeyFilter = C.napi_key_all_properties
	KeyWritable      KeyFilter = C.napi_key_writable
	KeyEnumerable    KeyFilter = C.napi_key_enumerable
	KeyConfigurable  KeyFilter = C.napi_key_configurable
	KeySkipStrings   KeyFilter = C.napi_key_skip_strings
	KeySkipSymbols   KeyFilter = C.napi_key_skip_symbols
)

// KeyConversion type
type KeyConversion C.napi_key_conversion

const (
	KeyKeepNumbers      KeyConversion = C.napi_key_keep_numbers
	KeyNumbersToStrings KeyConversion = C.napi_key_numbers_to_strings
)
