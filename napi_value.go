package napi

import (
	"fmt"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

// internal use to dont expose ValueType
type value = ValueType

// ValueType defines an interface for types that can be represented as N-API values.
// It provides methods to retrieve the underlying napi.Value and napi.Env, as well as
// methods to access the environment and determine the N-API type of the value.
//
// Implementations of ValueType are expected to bridge Go values with their
// corresponding N-API representations, facilitating interoperability with Node.js.
//
// Methods:
//   - NapiValue(): Returns the underlying primitive napi.Value for N-API calls.
//   - NapiEnv(): Returns the associated napi.Env for N-API calls.
//   - Env(): Returns the environment as an EnvType.
//   - Type(): Returns the N-API type of the value, or an error if it cannot be determined.
type ValueType interface {
	Env() EnvType            // NAPI Env to NAPI call
	Type() (NapiType, error) // NAPI Type of value
	NapiValue() napi.Value   // Primitive value to NAPI call
	NapiEnv() napi.Env       // NAPI Env to NAPI call
}

// Global napi-go [ValueType] from [napi._Value]
type _Value struct {
	env     EnvType    // Env
	valueOf napi.Value // JS value
	extra   []any      // Extends value
}

// Typeof of Value
type NapiType int

// String returns the string representation of the NapiType.
// If the NapiType is known, it returns its name; otherwise,
// it returns a formatted string indicating an unknown type.
func (t NapiType) String() string {
	switch t {
	case TypeUndefined:
		return "undefined"
	case TypeNull:
		return "null"
	case TypeBoolean:
		return "boolean"
	case TypeNumber:
		return "number"
	case TypeBigInt:
		return "bigint"
	case TypeString:
		return "string"
	case TypeSymbol:
		return "symbol"
	case TypeObject:
		return "object"
	case TypeFunction:
		return "function"
	case TypeExternal:
		return "external"
	case TypeTypedArray:
		return "typedarray"
	case TypePromise:
		return "promise"
	case TypeDataView:
		return "daraview"
	case TypeBuffer:
		return "buffer"
	case TypeDate:
		return "date"
	case TypeArray:
		return "array"
	case TypeArrayBuffer:
		return "arraybuffer"
	case TypeError:
		return "error"
	case TypeUnkown:
		return "unknown"
	default:
		return fmt.Sprintf("unknown type %d", t)
	}
}

// napi-go stand types
const (
	TypeUnkown NapiType = iota
	TypeUndefined
	TypeNull
	TypeBoolean
	TypeNumber
	TypeBigInt
	TypeString
	TypeSymbol
	TypeObject
	TypeFunction
	TypeExternal
	TypeTypedArray
	TypePromise
	TypeDataView
	TypeBuffer
	TypeDate
	TypeArray
	TypeArrayBuffer
	TypeError
)

// Return [ValueType] from [napi.Value]
func N_APIValue(env EnvType, value napi.Value, extra ...any) ValueType {
	return &_Value{env: env, valueOf: value, extra: extra}
}

func (v *_Value) NapiValue() napi.Value { return v.valueOf }
func (v *_Value) NapiEnv() napi.Env     { return v.env.NapiValue() }
func (v *_Value) Env() EnvType          { return v.env }
func (v *_Value) Type() (NapiType, error) {
	isTypedArray, status := napi.IsTypedArray(v.NapiEnv(), v.NapiValue())
	if err := status.ToError(); err != nil {
		return TypeUnkown, err
	}
	isPromise, status := napi.IsPromise(v.NapiEnv(), v.NapiValue())
	if err := status.ToError(); err != nil {
		return TypeUnkown, err
	}
	isDataView, status := napi.IsDataView(v.NapiEnv(), v.NapiValue())
	if err := status.ToError(); err != nil {
		return TypeUnkown, err
	}
	isBuffer, status := napi.IsBuffer(v.NapiEnv(), v.NapiValue())
	if err := status.ToError(); err != nil {
		return TypeUnkown, err
	}
	isDate, status := napi.IsDate(v.NapiEnv(), v.NapiValue())
	if err := status.ToError(); err != nil {
		return TypeUnkown, err
	}
	isArray, status := napi.IsArray(v.NapiEnv(), v.NapiValue())
	if err := status.ToError(); err != nil {
		return TypeUnkown, err
	}
	isArrayBuffer, status := napi.IsArrayBuffer(v.NapiEnv(), v.NapiValue())
	if err := status.ToError(); err != nil {
		return TypeUnkown, err
	}
	isError, status := napi.IsError(v.NapiEnv(), v.NapiValue())
	if err := status.ToError(); err != nil {
		return TypeUnkown, err
	}
	isTypeof, status := napi.Typeof(v.NapiEnv(), v.NapiValue())
	if err := status.ToError(); err != nil {
		return TypeUnkown, err
	}

	switch {
	case isTypedArray:
		return TypeTypedArray, nil
	case isPromise:
		return TypePromise, nil
	case isDataView:
		return TypeDataView, nil
	case isBuffer:
		return TypeBuffer, nil
	case isDate:
		return TypeDate, nil
	case isArray:
		return TypeArray, nil
	case isArrayBuffer:
		return TypeArrayBuffer, nil
	case isError:
		return TypeError, nil
	case isTypeof == napi.ValueTypeUndefined:
		return TypeUndefined, nil
	case isTypeof == napi.ValueTypeNull:
		return TypeNull, nil
	case isTypeof == napi.ValueTypeBoolean:
		return TypeBoolean, nil
	case isTypeof == napi.ValueTypeNumber:
		return TypeNumber, nil
	case isTypeof == napi.ValueTypeString:
		return TypeString, nil
	case isTypeof == napi.ValueTypeSymbol:
		return TypeSymbol, nil
	case isTypeof == napi.ValueTypeObject:
		return TypeObject, nil
	case isTypeof == napi.ValueTypeFunction:
		return TypeFunction, nil
	case isTypeof == napi.ValueTypeExternal:
		return TypeExternal, nil
	case isTypeof == napi.ValueTypeBigint:
		return TypeBigInt, nil
	}
	return TypeUnkown, nil
}

// This API represents the invocation of the Strict Equality algorithm as defined in https://tc39.github.io/ecma262/#sec-strict-equality-comparison of the ECMAScript Language Specification.
func StrictEqual(env EnvType, lhs, rhs ValueType) (bool, error) {
	ok, status := napi.StrictEquals(env.NapiValue(), lhs.NapiValue(), rhs.NapiValue())
	return ok, status.ToError()
}

// Casts to another type of [ValueType], when the actual type is known or
// assumed.
//
// This conversion does NOT coerce the type. Calling any methods
// inappropriate for the actual value type.
func As[T ValueType](input ValueType) T {
	v, ok := input.(T)
	if ok {
		return v
	}

	rawValue := N_APIValue(input.Env(), input.NapiValue())
	switch any(*new(T)).(type) {
	case *String, String:
		rawValue = &String{value: rawValue}
		return rawValue.(T)
	case *Number, Number:
		rawValue = &Number{value: rawValue}
		return rawValue.(T)
	case *Bigint, Bigint:
		rawValue = &Bigint{value: rawValue}
		return rawValue.(T)
	case *Boolean, Boolean:
		rawValue = &Boolean{value: rawValue}
		return rawValue.(T)
	case *Object, Object:
		rawValue = &Object{value: rawValue}
		return rawValue.(T)
	case *Array, Array:
		rawValue = &Array{value: rawValue}
		return rawValue.(T)
	case *ArrayBuffer, ArrayBuffer:
		rawValue = &ArrayBuffer{value: rawValue}
		return rawValue.(T)
	case *Date, Date:
		rawValue = &Date{value: rawValue}
		return rawValue.(T)
	default:
		return rawValue.(T)
	}
}
