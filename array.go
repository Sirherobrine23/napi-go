package napi

import (
	"iter"
	"slices"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type Array struct{ *Value }

func CreateArray(env *Env) (*Array, error) {
	value, err := napi.MustValueErr(napi.CreateArray(env.NapiValue()))
	if err != nil {
		return nil, err
	}

	return &Array{
		Value: &Value{
			env:     env,
			valueOf: value,
			typeof:  napi.MustValue(napi.Typeof(env.NapiValue(), value)),
		},
	}, nil
}

func (arr *Array) Length() (int, error) {
	return napi.MustValueErr(napi.GetArrayLength(arr.NapiEnv(), arr.NapiValue()))
}

func (arr *Array) Seq() iter.Seq[*Value] {
	return func(yield func(*Value) bool) {
		for index := range napi.MustValue(napi.GetArrayLength(arr.NapiEnv(), arr.NapiValue())) {
			value, err := arr.Get(index)
			if err != nil {
				return
			} else if !yield(value) {
				return
			}
		}
	}
}

func (arr *Array) Values() []*Value { return slices.Collect(arr.Seq()) }

func (arr *Array) Get(index int) (*Value, error) {
	value, err := napi.MustValueErr(napi.GetElement(arr.NapiEnv(), arr.NapiValue(), index))
	if err != nil {
		return nil, err
	}
	return &Value{
		env:     arr.env,
		valueOf: value,
		typeof:  napi.MustValue(napi.Typeof(arr.NapiEnv(), value)),
	}, nil
}

func (arr *Array) Set(index int, value *Value) error {
	status := napi.SetElement(arr.NapiEnv(), arr.NapiValue(), index, value.NapiValue())
	if status != napi.StatusOK {
		return napi.StatusError(status)
	}
	return nil
}

func (arr *Array) Delete(index int) (bool, error) {
	return napi.MustValueErr(napi.DeleteElement(arr.NapiEnv(), arr.NapiValue(), index))
}

func (arr *Array) Append(value *Value) error {
	index, err := napi.MustValueErr(napi.GetArrayLength(arr.NapiEnv(), arr.NapiValue()))
	if err != nil {
		return err
	}
	return arr.Set(index, value)
}
