package napi

import (
	"iter"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type Array struct{ *Value }

func (arr *Array) Length() (int, error) {
	return napi.MustValueErr(napi.GetArrayLength(arr.NapiEnv(), arr.NapiValue()))
}

func (arr *Array) Seq() iter.Seq[*Value] {
	return func(yield func(*Value) bool) {
		for index := range napi.MustValue(napi.GetArrayLength(arr.NapiEnv(), arr.NapiValue())) {
			value := napi.MustValue(napi.GetElement(arr.NapiEnv(), arr.NapiValue(), index))
			if !yield(&Value{env: arr.env, typeof: napi.MustValue(napi.Typeof(arr.NapiEnv(), value)), valueOf: value}) {
				return
			}
		}
	}
}
