package napi

import (
	"iter"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type Object struct{ value }

func CreateObject(env EnvType) (*Object, error) {
	value, err := napi.MustValueErr(napi.CreateObject(env.NapiValue()))
	if err != nil {
		return nil, err
	}
	return &Object{FromValueNapi(env, value)}, nil
}

func (obj *Object) Has(name string) bool {
	if exist, status := napi.HasNamedProperty(obj.Env().NapiValue(), obj.NapiValue(), name); status == napi.StatusOK {
		return exist
	}
	return false
}

func (obj *Object) HasOwnProperty(name string) bool {
	v, err := CreateString(obj.Env(), name)
	if err != nil {
		return false
	}

	if exist, status := napi.HasOwnProperty(obj.Env().NapiValue(), obj.NapiValue(), v.NapiValue()); status == napi.StatusOK {
		return exist
	}
	return false
}

func (obj *Object) Set(name string, value ValueType) error {
	if status := napi.SetNamedProperty(obj.Env().NapiValue(), obj.NapiValue(), name, value.NapiValue()); status != napi.StatusOK {
		return napi.StatusError(status)
	}
	return nil
}

func (obj *Object) Delete(name string) error {
	v, err := CreateString(obj.Env(), name)
	if err != nil {
		return err
	}

	_, status := napi.DeleteProperty(obj.Env().NapiValue(), obj.NapiValue(), v.NapiValue())
	if status != napi.StatusOK {
		return napi.StatusError(status)
	}

	return nil
}

func (obj *Object) GetPropertyNames() (*Array, error) {
	array, err := napi.MustValueErr(napi.GetPropertyNames(obj.Env().NapiValue(), obj.NapiValue()))
	if err != nil {
		return nil, err
	}
	return &Array{FromValueNapi(obj.Env(), array)}, nil
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

func (obj *Object) Seq() iter.Seq2[string, ValueType] {
	arr, err := obj.GetPropertyNames()
	if err != nil {
		return nil
	}

	return func(yield func(string, ValueType) bool) {
		for nameValue := range arr.Seq() {
			name := nameValue.ToString().String()
			value := napi.MustValue(napi.GetNamedProperty(arr.NapiEnv(), arr.NapiValue(), name))
			if !yield(name, &Value{env: obj.Env(), valueOf: value}) {
				return
			}
		}
	}
}

func (obj *Object) Get(name string) (ValueType, error) {
	v, err := CreateString(obj.Env(), name)
	if err != nil {
		return nil, err
	}
	value, err := napi.MustValueErr(napi.GetProperty(obj.NapiEnv(), obj.NapiValue(), v.NapiValue()))
	if err != nil {
		return nil, err
	}
	return FromValueNapi(obj.Env(), value), nil
}
