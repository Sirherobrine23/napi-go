package napi

import (
	"iter"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type Object struct{ *Value }

func (obj *Object) Has(name string) bool {
	if exist, status := napi.HasNamedProperty(obj.env.NapiValue(), obj.NapiValue(), name); status == napi.StatusOK {
		return exist
	}
	return false
}

func (obj *Object) HasOwnProperty(name string) bool {
	v, err := FromString(obj.env, name)
	if err != nil {
		return false
	}

	if exist, status := napi.HasOwnProperty(obj.env.NapiValue(), obj.NapiValue(), v.NapiValue()); status == napi.StatusOK {
		return exist
	}
	return false
}

func (obj *Object) Set(name string, value *Value) error {
	if status := napi.SetNamedProperty(obj.env.NapiValue(), obj.NapiValue(), name, value.NapiValue()); status != napi.StatusOK {
		return napi.StatusError(status)
	}
	return nil
}

func (obj *Object) Delete(name string) error {
	v, err := FromString(obj.env, name)
	if err != nil {
		return err
	}

	_, status := napi.DeleteProperty(obj.env.NapiValue(), obj.NapiValue(), v.NapiValue())
	if status != napi.StatusOK {
		return napi.StatusError(status)
	}

	return nil
}

func (obj *Object) GetPropertyNames() (*Array, error) {
	array, status := napi.GetPropertyNames(obj.env.NapiValue(), obj.NapiValue())
	if status != napi.StatusOK {
		return nil, napi.StatusError(status)
	}
	return &Array{Value: &Value{env: obj.env, typeof: 0, valueOf: array}}, nil
}

func (obj *Object) SeqGetPropertyNames() iter.Seq2[string, error] {
	arr, err := obj.GetPropertyNames()
	if err != nil {
		return func(yield func(string, error) bool) { yield("", err) }
	}

	sliceSize := napi.MustValue(napi.GetArrayLength(arr.NapiEnv(), arr.NapiValue()))
	return func(yield func(string, error) bool) {
		for index := range sliceSize {
			if !yield(napi.MustValueErr(napi.GetValueStringUtf8(arr.NapiEnv(), napi.MustValue(napi.GetElement(arr.NapiEnv(), arr.NapiValue(), index))))) {
				return
			}
		}
	}
}

func (obj *Object) Seq() iter.Seq2[string, *Value] {
	arr, err := obj.GetPropertyNames()
	if err != nil {
		return nil
	}

	return func(yield func(string, *Value) bool) {
		for nameValue := range arr.Seq() {
			name := nameValue.ToString().String()
			value := napi.MustValue(napi.GetNamedProperty(arr.NapiEnv(), arr.NapiValue(), name))
			if !yield(name, &Value{env: obj.env, valueOf: value}) {
				return
			}
		}
	}
}
