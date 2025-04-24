package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

type (
	Env   struct{ env napi.Env }
	Value struct {
		env     *Env
		typeof  napi.ValueType
		valueOf napi.Value
	}
)

// Return internal Napi Value point
func (e *Env) NapiValue() napi.Env { return e.env }

func (e *Env) Global() *Object {
	return &Object{
		Value: &Value{
			env:     e,
			valueOf: napi.MustValue(napi.GetGlobal(e.env)),
			typeof:  napi.ValueTypeObject,
		},
	}
}

func (e *Env) Undefined() *Value {
	return &Value{
		env:     e,
		typeof:  napi.ValueTypeUndefined,
		valueOf: napi.MustValue(napi.GetUndefined(e.env)),
	}
}

func (e *Env) Null() *Value {
	return &Value{
		env:     e,
		typeof:  napi.ValueTypeUndefined,
		valueOf: napi.MustValue(napi.GetNull(e.env)),
	}
}

// Return internal Napi Value point
func (v *Value) NapiValue() napi.Value { return v.valueOf }
func (v *Value) NapiEnv() napi.Env     { return v.env.env }

func (v *Value) IsUndefined() bool { return v.typeof == napi.ValueTypeUndefined }
func (v *Value) IsNull() bool      { return v.typeof == napi.ValueTypeNull }
func (v *Value) IsBoolean() bool   { return v.typeof == napi.ValueTypeBoolean }
func (v *Value) IsNumber() bool    { return v.typeof == napi.ValueTypeNumber }
func (v *Value) IsBigInt() bool    { return v.typeof == napi.ValueTypeBigint }
func (v *Value) IsString() bool    { return v.typeof == napi.ValueTypeString }
func (v *Value) IsSymbol() bool    { return v.typeof == napi.ValueTypeSymbol }
func (v *Value) IsObject() bool    { return v.typeof == napi.ValueTypeObject }
func (v *Value) IsFunction() bool  { return v.typeof == napi.ValueTypeFunction }
func (v *Value) IsExternal() bool  { return v.typeof == napi.ValueTypeExternal }

func (v *Value) IsDate() bool        { return napi.MustValue(napi.IsDate(v.env.env, v.valueOf)) }
func (v *Value) IsArray() bool       { return napi.MustValue(napi.IsArray(v.env.env, v.valueOf)) }
func (v *Value) IsArrayBuffer() bool { return napi.MustValue(napi.IsArrayBuffer(v.env.env, v.valueOf)) }
func (v *Value) IsTypedArray() bool  { return napi.MustValue(napi.IsTypedArray(v.env.env, v.valueOf)) }
func (v *Value) IsPromise() bool     { return napi.MustValue(napi.IsPromise(v.env.env, v.valueOf)) }
func (v *Value) IsDataView() bool    { return napi.MustValue(napi.IsDataView(v.env.env, v.valueOf)) }
func (v *Value) IsBuffer() bool      { return napi.MustValue(napi.IsBuffer(v.env.env, v.valueOf)) }

func (v *Value) ToBoolean() *Boolean { return &Boolean{Value: v} }
func (v *Value) ToNumber() *Number   { return &Number{Value: v} }
func (v *Value) ToString() *String   { return &String{Value: v} }
func (v *Value) ToObject() *Object   { return &Object{Value: v} }
