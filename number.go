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

// Float returns the float64 representation of the Number.
// It retrieves the underlying value from the N-API environment and value handle.
// If the conversion fails, an error is returned.
func (num *Number) Float() (float64, error) {
	return mustValueErr(napi.GetValueDouble(num.NapiEnv(), num.NapiValue()))
}

// Int returns the int64 representation of the Number.
// It calls napi.GetValueInt64 using the underlying NapiEnv and NapiValue.
// If the conversion fails, an error is returned.
func (num *Number) Int() (int64, error) {
	return mustValueErr(napi.GetValueInt64(num.NapiEnv(), num.NapiValue()))
}

// Uint32 retrieves the value of the Number as a uint32.
// It returns the uint32 representation of the Number and an error if the conversion fails.
func (num *Number) Uint32() (uint32, error) {
	return mustValueErr(napi.GetValueUint32(num.NapiEnv(), num.NapiValue()))
}

// Int32 retrieves the value of the Number as an int32.
// It returns the int32 representation of the Number and an error if the conversion fails.
func (num *Number) Int32() (int32, error) {
	return mustValueErr(napi.GetValueInt32(num.NapiEnv(), num.NapiValue()))
}

// Int64 returns the value of the Bigint as an int64 along with an error if the conversion fails.
// It retrieves the int64 representation of the underlying N-API BigInt value.
// If the value cannot be represented as an int64, an error is returned.
func (big *Bigint) Int64() (int64, error) {
	return mustValueErr2(napi.GetValueBigIntInt64(big.NapiEnv(), big.NapiValue()))
}

// Uint64 returns the value of the Bigint as a uint64 along with an error if the conversion fails.
// It retrieves the underlying BigInt value from the N-API environment and attempts to convert it to a uint64.
// If the value cannot be represented as a uint64 or if an error occurs during retrieval, an error is returned.
func (big *Bigint) Uint64() (uint64, error) {
	return mustValueErr2(napi.GetValueBigIntUint64(big.NapiEnv(), big.NapiValue()))
}

// CreateBigint creates a new Bigint value in the given N-API environment from the provided int64 or uint64 value.
// The function is generic and accepts either int64 or uint64 as the input type.
// It returns a pointer to a Bigint and an error if the creation fails.
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

	return ToBigint(N_APIValue(env, value)), nil
}

// CreateNumber creates a new JavaScript Number object from a Go numeric value of type T.
// The function supports various integer and floating-point types, including int, uint, int8, uint8,
// int16, uint16, int32, uint32, int64, uint64, float32, and float64. It converts the provided Go
// value to the appropriate JavaScript number representation using the N-API environment.
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
