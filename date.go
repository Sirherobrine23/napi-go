package napi

import (
	"time"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type Date struct{ value }

// Convert [ValueType] to [*Date]
func ToDate(o ValueType) *Date { return &Date{o} }

// CreateDate creates a new JavaScript Date object in the given N-API environment
// using the provided Go time.Time value. It returns a pointer to a Date wrapper
// or an error if the creation fails.
func CreateDate(env EnvType, t time.Time) (*Date, error) {
	value, err := mustValueErr(napi.CreateDate(env.NapiValue(), float64(t.UnixMilli())))
	if err != nil {
		return nil, err
	}
	return &Date{value: &_Value{env: env, valueOf: value}}, nil
}

// Time returns the Go time.Time representation of the Date value.
// It retrieves the date value from the underlying N-API environment,
// converts it to a Unix millisecond timestamp, and constructs a time.Time object.
// If an error occurs during value retrieval or conversion, it is returned.
func (d Date) Time() (time.Time, error) {
	timeFloat, err := mustValueErr(napi.GetDateValue(d.NapiEnv(), d.NapiValue()))
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(int64(timeFloat)), nil
}
