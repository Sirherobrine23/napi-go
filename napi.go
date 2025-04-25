package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

type (
	ValueType interface {
		NapiValue() napi.Value // Primitive value to NAPI call
		NapiEnv() napi.Env     // NAPI Env to NAPI call

		Env() EnvType // NAPI Env to NAPI call

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

		ToBoolean() *Boolean // Convert to Boolean value
		ToNumber() *Number   // Convert to Number value
		ToObject() *Object   // Convert to Object value
		ToString() *String   // Convert to String value
	}

	EnvType interface {
		NapiValue() napi.Env // Primitive value to NAPI call
		Global() *Object
		MustUndefined() ValueType
		Null() ValueType
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

func (e *Env) Global() *Object {
	return &Object{FromValueNapi(e, napi.MustValue(napi.GetGlobal(e.env)))}
}

func (e *Env) MustUndefined() ValueType {
	return FromValueNapi(e, napi.MustValue(napi.GetUndefined(e.env)))
}

func (e *Env) Null() ValueType {
	return FromValueNapi(e, napi.MustValue(napi.GetNull(e.env)))
}

// Return internal Napi Value point
func (v *Value) NapiValue() napi.Value { return v.valueOf }
func (v *Value) NapiEnv() napi.Env     { return v.env.NapiValue() }
func (v *Value) Env() EnvType          { return v.env }

func (v *Value) IsUndefined() bool {
	return napi.MustValue(napi.Typeof(v.NapiEnv(), v.valueOf)) == napi.ValueTypeUndefined
}
func (v *Value) IsNull() bool {
	return napi.MustValue(napi.Typeof(v.NapiEnv(), v.valueOf)) == napi.ValueTypeNull
}
func (v *Value) IsBoolean() bool {
	return napi.MustValue(napi.Typeof(v.NapiEnv(), v.valueOf)) == napi.ValueTypeBoolean
}
func (v *Value) IsNumber() bool {
	return napi.MustValue(napi.Typeof(v.NapiEnv(), v.valueOf)) == napi.ValueTypeNumber
}
func (v *Value) IsBigInt() bool {
	return napi.MustValue(napi.Typeof(v.NapiEnv(), v.valueOf)) == napi.ValueTypeBigint
}
func (v *Value) IsString() bool {
	return napi.MustValue(napi.Typeof(v.NapiEnv(), v.valueOf)) == napi.ValueTypeString
}
func (v *Value) IsSymbol() bool {
	return napi.MustValue(napi.Typeof(v.NapiEnv(), v.valueOf)) == napi.ValueTypeSymbol
}
func (v *Value) IsObject() bool {
	return napi.MustValue(napi.Typeof(v.NapiEnv(), v.valueOf)) == napi.ValueTypeObject
}
func (v *Value) IsFunction() bool {
	return napi.MustValue(napi.Typeof(v.NapiEnv(), v.valueOf)) == napi.ValueTypeFunction
}
func (v *Value) IsExternal() bool {
	return napi.MustValue(napi.Typeof(v.NapiEnv(), v.valueOf)) == napi.ValueTypeExternal
}

func (v *Value) IsTypedArray() bool { return napi.MustValue(napi.IsTypedArray(v.NapiEnv(), v.valueOf)) }
func (v *Value) IsPromise() bool    { return napi.MustValue(napi.IsPromise(v.NapiEnv(), v.valueOf)) }
func (v *Value) IsDataView() bool   { return napi.MustValue(napi.IsDataView(v.NapiEnv(), v.valueOf)) }
func (v *Value) IsBuffer() bool     { return napi.MustValue(napi.IsBuffer(v.NapiEnv(), v.valueOf)) }
func (v *Value) IsDate() bool       { return napi.MustValue(napi.IsDate(v.NapiEnv(), v.valueOf)) }
func (v *Value) IsArray() bool      { return napi.MustValue(napi.IsArray(v.NapiEnv(), v.valueOf)) }
func (v *Value) IsArrayBuffer() bool {
	return napi.MustValue(napi.IsArrayBuffer(v.NapiEnv(), v.valueOf))
}

func (v *Value) ToBoolean() *Boolean { return &Boolean{value: v} }
func (v *Value) ToNumber() *Number   { return &Number{value: v} }
func (v *Value) ToString() *String   { return &String{value: v} }
func (v *Value) ToObject() *Object   { return &Object{value: v} }
