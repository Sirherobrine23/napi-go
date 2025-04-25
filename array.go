package napi

import (
	"iter"
	"slices"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type Array struct{ value }

func CreateArray(env EnvType) (*Array, error) {
	value, err := napi.MustValueErr(napi.CreateArray(env.NapiValue()))
	if err != nil {
		return nil, err
	}

	return &Array{value: FromValueNapi(env, value)}, nil
}

func (arr *Array) Length() (int, error) {
	return napi.MustValueErr(napi.GetArrayLength(arr.NapiEnv(), arr.NapiValue()))
}

func (arr *Array) Seq() iter.Seq[ValueType] {
	return func(yield func(ValueType) bool) {
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

func (arr *Array) Values() []ValueType { return slices.Collect(arr.Seq()) }

func (arr *Array) Get(index int) (ValueType, error) {
	value, err := napi.MustValueErr(napi.GetElement(arr.NapiEnv(), arr.NapiValue(), index))
	if err != nil {
		return nil, err
	}
	return FromValueNapi(arr.Env(), value), nil
}

func (arr *Array) Set(index int, value ValueType) error {
	return napi.SingleMustValueErr(napi.SetElement(arr.NapiEnv(), arr.NapiValue(), index, value.NapiValue()))
}

func (arr *Array) Delete(index int) (bool, error) {
	return napi.MustValueErr(napi.DeleteElement(arr.NapiEnv(), arr.NapiValue(), index))
}

func (arr *Array) Append(value ValueType) error {
	index, err := napi.MustValueErr(napi.GetArrayLength(arr.NapiEnv(), arr.NapiValue()))
	if err == nil {
		err = arr.Set(index, value)
	}
	return err
}
