package napi

import (
	"fmt"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type ValueType interface {
	NapiValue() napi.Value // Primitive value to NAPI call
	NapiEnv() napi.Env     // NAPI Env to NAPI call

	Env() EnvType            // NAPI Env to NAPI call
	Type() (NapiType, error) // NAPI Type of value
}

type (
	value = ValueType // to dont expose to external structs

	// Generic type to NAPI value
	Value struct {
		env     EnvType
		valueOf napi.Value
	}

	NapiType int // Return typeof of Value
)

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

var napiTypeNames = map[NapiType]string{
	TypeUnkown:      "Unknown",
	TypeUndefined:   "Undefined",
	TypeNull:        "Null",
	TypeBoolean:     "Boolean",
	TypeNumber:      "Number",
	TypeBigInt:      "BigInt",
	TypeString:      "String",
	TypeSymbol:      "Symbol",
	TypeObject:      "Object",
	TypeFunction:    "Function",
	TypeExternal:    "External",
	TypeTypedArray:  "TypedArray",
	TypePromise:     "Promise",
	TypeDataView:    "DaraView",
	TypeBuffer:      "Buffer",
	TypeDate:        "Date",
	TypeArray:       "Array",
	TypeArrayBuffer: "ArrayBuffer",
	TypeError:       "Error",
}

// Return [ValueType] from [napi.Value]
func N_APIValue(env EnvType, value napi.Value) ValueType {
	return &Value{env: env, valueOf: value}
}

func (v *Value) NapiValue() napi.Value { return v.valueOf }
func (v *Value) NapiEnv() napi.Env     { return v.env.NapiValue() }
func (v *Value) Env() EnvType          { return v.env }

func (v *Value) Type() (NapiType, error) {
	isTypedArray, err := mustValueErr(napi.IsTypedArray(v.NapiEnv(), v.NapiValue()))
	if err != nil {
		return TypeUnkown, err
	}
	isPromise, err := mustValueErr(napi.IsPromise(v.NapiEnv(), v.NapiValue()))
	if err != nil {
		return TypeUnkown, err
	}
	isDataView, err := mustValueErr(napi.IsDataView(v.NapiEnv(), v.NapiValue()))
	if err != nil {
		return TypeUnkown, err
	}
	isBuffer, err := mustValueErr(napi.IsBuffer(v.NapiEnv(), v.NapiValue()))
	if err != nil {
		return TypeUnkown, err
	}
	isDate, err := mustValueErr(napi.IsDate(v.NapiEnv(), v.NapiValue()))
	if err != nil {
		return TypeUnkown, err
	}
	isArray, err := mustValueErr(napi.IsArray(v.NapiEnv(), v.NapiValue()))
	if err != nil {
		return TypeUnkown, err
	}
	isArrayBuffer, err := mustValueErr(napi.IsArrayBuffer(v.NapiEnv(), v.NapiValue()))
	if err != nil {
		return TypeUnkown, err
	}
	isError, err := mustValueErr(napi.IsError(v.NapiEnv(), v.NapiValue()))
	if err != nil {
		return TypeUnkown, err
	}
	isTypeof, err := mustValueErr(napi.Typeof(v.NapiEnv(), v.NapiValue()))
	if err != nil {
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

func (t NapiType) String() string {
	if name, ok := napiTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("Unknown NapiType %d", t)
}
