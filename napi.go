package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

type (
	ValueType interface {
		NapiValue() napi.Value // Primitive value to NAPI call
		NapiEnv() napi.Env     // NAPI Env to NAPI call

		Env() EnvType   // NAPI Env to NAPI call
		Type() NapiType // NAPI Type of value

		IsArray() bool
		IsArrayBuffer() bool
		IsBigInt() bool
		IsBoolean() bool
		IsBuffer() bool
		IsDataView() bool
		IsDate() bool
		IsExternal() bool
		IsFunction() bool
		IsNull() bool
		IsNumber() bool
		IsObject() bool
		IsPromise() bool
		IsString() bool
		IsSymbol() bool
		IsTypedArray() bool
		IsUndefined() bool
		IsError() bool

		ToBoolean() *Boolean // Convert to Boolean value
		ToNumber() *Number   // Convert to Number value
		ToObject() *Object   // Convert to Object value
		ToString() *String   // Convert to String value
	}

	EnvType interface {
		NapiValue() napi.Env // Primitive value to NAPI call
		Global() (*Object, error)
		Undefined() (ValueType, error)
		Null() (ValueType, error)
		MustGlobal() *Object
		MustUndefined() ValueType
		MustNull() ValueType
	}

	value = ValueType // to dont expose to external structs

	Env   struct{ env napi.Env }
	Value struct {
		env     EnvType
		valueOf napi.Value
	}
)

func FromEnvNapi(env napi.Env) EnvType { return &Env{env} }
func FromValueNapi(env EnvType, value napi.Value) ValueType {
	return &Value{env: env, valueOf: value}
}

// Return internal Napi Value point
func (e *Env) NapiValue() napi.Env { return e.env }

func (e *Env) Global() (*Object, error) {
	napiValue, err := napi.MustValueErr(napi.GetGlobal(e.env))
	if err != nil {
		return nil, err
	}
	return FromValueNapi(e, napiValue).ToObject(), nil
}
func (e *Env) MustGlobal() *Object {
	value, err := e.Global()
	if err != nil {
		panic(err)
	}
	return value
}

func (e *Env) Undefined() (ValueType, error) {
	napiValue, err := napi.MustValueErr(napi.GetUndefined(e.env))
	if err != nil {
		return nil, err
	}
	return FromValueNapi(e, napiValue), nil
}
func (e *Env) MustUndefined() ValueType {
	value, err := e.Undefined()
	if err != nil {
		panic(err)
	}
	return value
}

func (e *Env) Null() (ValueType, error) {
	napiValue, err := napi.MustValueErr(napi.GetNull(e.env))
	if err != nil {
		return nil, err
	}
	return FromValueNapi(e, napiValue), nil
}
func (e *Env) MustNull() ValueType {
	value, err := e.Null()
	if err != nil {
		panic(err)
	}
	return value
}

// Return internal Napi Value point
func (v *Value) NapiValue() napi.Value { return v.valueOf }
func (v *Value) NapiEnv() napi.Env     { return v.env.NapiValue() }
func (v *Value) Env() EnvType          { return v.env }

func (v *Value) IsUndefined() bool { return v.Type() == TypeUndefined }
func (v *Value) IsNull() bool      { return v.Type() == TypeNull }
func (v *Value) IsBoolean() bool   { return v.Type() == TypeBoolean }
func (v *Value) IsNumber() bool    { return v.Type() == TypeNumber }
func (v *Value) IsBigInt() bool    { return v.Type() == TypeBigInt }
func (v *Value) IsString() bool    { return v.Type() == TypeString }
func (v *Value) IsSymbol() bool    { return v.Type() == TypeSymbol }
func (v *Value) IsObject() bool    { return v.Type() == TypeObject }
func (v *Value) IsFunction() bool  { return v.Type() == TypeFunction }
func (v *Value) IsExternal() bool  { return v.Type() == TypeExternal }

func (v *Value) IsTypedArray() bool { return napi.MustValue(napi.IsTypedArray(v.NapiEnv(), v.valueOf)) }
func (v *Value) IsPromise() bool    { return napi.MustValue(napi.IsPromise(v.NapiEnv(), v.valueOf)) }
func (v *Value) IsDataView() bool   { return napi.MustValue(napi.IsDataView(v.NapiEnv(), v.valueOf)) }
func (v *Value) IsBuffer() bool     { return napi.MustValue(napi.IsBuffer(v.NapiEnv(), v.valueOf)) }
func (v *Value) IsDate() bool       { return napi.MustValue(napi.IsDate(v.NapiEnv(), v.valueOf)) }
func (v *Value) IsArray() bool      { return napi.MustValue(napi.IsArray(v.NapiEnv(), v.valueOf)) }
func (v *Value) IsArrayBuffer() bool {
	return napi.MustValue(napi.IsArrayBuffer(v.NapiEnv(), v.valueOf))
}
func (v *Value) IsError() bool {
	return napi.MustValue(napi.IsError(v.NapiEnv(), v.valueOf))
}

func (v *Value) ToBoolean() *Boolean { return &Boolean{value: v} }
func (v *Value) ToNumber() *Number   { return &Number{value: v} }
func (v *Value) ToString() *String   { return &String{value: v} }
func (v *Value) ToObject() *Object   { return &Object{value: v} }

func (v *Value) Type() NapiType {
	switch {
	case v.IsDate():
		return TypeDate
	case v.IsArrayBuffer():
		return TypeArrayBuffer
	case v.IsArray():
		return TypeArray
	case v.IsBuffer():
		return TypeBuffer
	case v.IsDataView():
		return TypeDataView
	case v.IsTypedArray():
		return TypeTypedArray
	case v.IsPromise():
		return TypePromise
	case v.IsExternal():
		return TypeExternal
	case v.IsFunction():
		return TypeFunction
	case v.IsObject():
		return TypeObject
	case v.IsSymbol():
		return TypeSymbol
	case v.IsString():
		return TypeString
	case v.IsBigInt():
		return TypeBigInt
	case v.IsNumber():
		return TypeNumber
	case v.IsBoolean():
		return TypeBoolean
	case v.IsNull():
		return TypeNull
	case v.IsUndefined():
		return TypeUndefined
	default:
		return TypeUnkown
	}
}
