package napi

import (
	"fmt"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type Number struct{ value }
type Bigint struct{ value }

// Convert [ValueType] to [*Number]
func ToNumber(o ValueType) *Number { return &Number{o} }

// Convert [ValueType] to [*Bigint]
func ToBigint(o ValueType) *Bigint { return &Bigint{o} }

func (num *Number) Float() (float64, error) {
	return mustValueErr(napi.GetValueDouble(num.NapiEnv(), num.NapiValue()))
}

func (num *Number) Int() (int64, error) {
	return mustValueErr(napi.GetValueInt64(num.NapiEnv(), num.NapiValue()))
}

func (num *Number) Uint32() (uint32, error) {
	return mustValueErr(napi.GetValueUint32(num.NapiEnv(), num.NapiValue()))
}

func (num *Number) Int32() (int32, error) {
	return mustValueErr(napi.GetValueInt32(num.NapiEnv(), num.NapiValue()))
}

func (big *Bigint) Int64() (int64, error) {
	return mustValueErr2(napi.GetValueBigIntInt64(big.NapiEnv(), big.NapiValue()))
}
func (big *Bigint) Uint64() (uint64, error) {
	return mustValueErr2(napi.GetValueBigIntUint64(big.NapiEnv(), big.NapiValue()))
}

func CreateBigint[T int64 | uint64](env EnvType, valueOf T) (*Bigint, error) {
	var value napi.Value
	var err error
	switch v := any(valueOf).(type) {
	case int64:
		if value, err = mustValueErr(napi.CreateBigIntInt64(env.NapiValue(), v)); err != nil {
			return nil, err
		}
	case uint64:
		if value, err = mustValueErr(napi.CreateBigIntUint64(env.NapiValue(), v)); err != nil {
			return nil, err
		}
	}

	return &Bigint{value: &Value{env: env, valueOf: value}}, nil
}

func CreateNumber[T ~int | ~uint | ~int8 | ~uint8 | ~int16 | ~uint16 | ~int32 | ~uint32 | ~int64 | ~uint64 | ~float32 | ~float64](env EnvType, n T) (*Number, error) {
	var value napi.Value
	var err error
	switch v := any(n).(type) {
	case int:
		if value, err = mustValueErr(napi.CreateInt64(env.NapiValue(), int64(v))); err != nil {
			return nil, err
		}
	case uint:
		if value, err = mustValueErr(napi.CreateInt64(env.NapiValue(), int64(v))); err != nil {
			return nil, err
		}
	case int8:
		if value, err = mustValueErr(napi.CreateInt64(env.NapiValue(), int64(v))); err != nil {
			return nil, err
		}
	case uint8:
		if value, err = mustValueErr(napi.CreateInt64(env.NapiValue(), int64(v))); err != nil {
			return nil, err
		}
	case int16:
		if value, err = mustValueErr(napi.CreateInt64(env.NapiValue(), int64(v))); err != nil {
			return nil, err
		}
	case uint16:
		if value, err = mustValueErr(napi.CreateInt64(env.NapiValue(), int64(v))); err != nil {
			return nil, err
		}
	case int32:
		if value, err = mustValueErr(napi.CreateInt32(env.NapiValue(), v)); err != nil {
			return nil, err
		}
	case uint32:
		if value, err = mustValueErr(napi.CreateUint32(env.NapiValue(), v)); err != nil {
			return nil, err
		}
	case int64:
		if value, err = mustValueErr(napi.CreateInt64(env.NapiValue(), v)); err != nil {
			return nil, err
		}
	case uint64:
		if value, err = mustValueErr(napi.CreateInt64(env.NapiValue(), int64(v))); err != nil {
			return nil, err
		}
	case float32:
		if value, err = mustValueErr(napi.CreateDouble(env.NapiValue(), float64(v))); err != nil {
			return nil, err
		}
	case float64:
		if value, err = mustValueErr(napi.CreateDouble(env.NapiValue(), v)); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid number type")
	}
	return ToNumber(N_APIValue(env, value)), err
}
