package napi

import "fmt"

type NapiType int

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
}

func (t NapiType) String() string {
	if name, ok := napiTypeNames[t]; ok {
		return name
	}
	return fmt.Sprintf("Unknown NapiType %d", t)
}
