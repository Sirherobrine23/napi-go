package napi

import (
	"iter"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type Object struct{ value }

// Convert ValueType to [*Object]
func ToObject(o ValueType) *Object { return &Object{o} }

// Create [*Object]
func CreateObject(env EnvType) (*Object, error) {
	napiValue, err := mustValueErr(napi.CreateObject(env.NapiValue()))
	if err != nil {
		return nil, err
	}
	return ToObject(N_APIValue(env, napiValue)), nil
}

// Check if exists named property.
func (obj *Object) Has(name string) (bool, error) {
	return mustValueErr(napi.HasNamedProperty(obj.NapiEnv(), obj.NapiValue(), name))
}

// Checks whether a own property is present.
func (obj *Object) HasOwnProperty(key ValueType) (bool, error) {
	return mustValueErr(napi.HasOwnProperty(obj.NapiEnv(), obj.NapiValue(), key.NapiValue()))
}

// Checks whether a own property is present.
func (obj *Object) HasOwnPropertyString(keyString string) (bool, error) {
	napiString, err := CreateString(obj.Env(), keyString)
	if err != nil {
		return false, err
	}
	return obj.HasOwnProperty(napiString)
}

// Gets a property.
func (obj *Object) Get(key string) (ValueType, error) {
	keyValue, err := CreateString(obj.Env(), key)
	if err != nil {
		return nil, err
	}
	return obj.GetWithValue(keyValue)
}

// Gets a property.
func (obj *Object) GetWithValue(key ValueType) (ValueType, error) {
	napiValue, err := mustValueErr(napi.GetProperty(obj.Env().NapiValue(), obj.NapiValue(), key.NapiValue()))
	if err != nil {
		return nil, err
	}
	return N_APIValue(obj.Env(), napiValue), nil
}

// Sets a property.
func (obj *Object) Set(key string, value ValueType) error {
	keyValue, err := CreateString(obj.Env(), key)
	if err != nil {
		return err
	}
	return obj.SetWithValue(keyValue, value)
}

// Sets a property.
func (obj *Object) SetWithValue(key, value ValueType) error {
	return singleMustValueErr(napi.SetProperty(obj.NapiEnv(), obj.NapiValue(), key.NapiValue(), value.NapiValue()))
}

// Delete property.
func (obj *Object) Delete(key string) (bool, error) {
	keyValue, err := CreateString(obj.Env(), key)
	if err != nil {
		return false, err
	}
	return obj.DeleteWithValue(keyValue)
}

// Delete property.
func (obj *Object) DeleteWithValue(key ValueType) (bool, error) {
	return mustValueErr(napi.DeleteProperty(obj.NapiEnv(), obj.NapiValue(), key.NapiValue()))
}

// Get all property names.
func (obj *Object) GetPropertyNames() (*Array, error) {
	arrValue, err := mustValueErr(napi.GetPropertyNames(obj.NapiEnv(), obj.NapiValue()))
	if err != nil {
		return nil, err
	}
	return ToArray(N_APIValue(obj.Env(), arrValue)), nil
}

// Checks if an object is an instance created by a constructor function,
// this is equivalent to the JavaScript `instanceof` operator.
func (obj *Object) InstanceOf(value ValueType) (bool, error) {
	return mustValueErr(napi.InstanceOf(obj.NapiEnv(), obj.NapiValue(), value.NapiValue()))
}

// Freeze object.
func (obj *Object) Freeze() error {
	return singleMustValueErr(napi.ObjectFreeze(obj.NapiEnv(), obj.NapiValue()))
}

func (obj *Object) Seal() error {
	return singleMustValueErr(napi.ObjectSeal(obj.NapiEnv(), obj.NapiValue()))
}

func (obj *Object) Seq() iter.Seq2[string, ValueType] {
	keys, err := obj.GetPropertyNames()
	if err != nil {
		panic(err)
	}
	return func(yield func(string, ValueType) bool) {
		for key := range keys.Seq() {
			value, err := obj.GetWithValue(key)
			if err != nil {
				panic(err)
			}

			keyName, err := ToString(key).Utf8Value()
			if err != nil {
				panic(err)
			}

			if !yield(keyName, value) {
				return
			}
		}
	}
}
