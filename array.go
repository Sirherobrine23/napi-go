package napi

import (
	"iter"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type Array struct{ value }

// Convert [ValueType] to [*Array].
func ToArray(o ValueType) *Array { return &Array{o} }

// Create Array.
func CreateArray(env EnvType, size ...int) (*Array, error) {
	sizeOf := 0
	if len(size) > 0 {
		sizeOf = size[0]
	}
	napiValue, err := napi.Value(nil), error(nil)
	if sizeOf == 0 {
		napiValue, err = mustValueErr(napi.CreateArray(env.NapiValue()))
	} else {
		napiValue, err = mustValueErr(napi.CreateArrayWithLength(env.NapiValue(), sizeOf))
	}
	// Check error exists
	if err != nil {
		return nil, err
	}
	return ToArray(N_APIValue(env, napiValue)), nil
}

// Get array length.
func (arr *Array) Length() (int, error) {
	return mustValueErr(napi.GetArrayLength(arr.NapiEnv(), arr.NapiValue()))
}

// Delete index elemente from array.
func (arr *Array) Delete(index int) (bool, error) {
	return mustValueErr(napi.DeleteElement(arr.NapiEnv(), arr.NapiValue(), index))
}

// Set value in index
func (arr *Array) Set(index int, value ValueType) error {
	return singleMustValueErr(napi.SetElement(arr.NapiEnv(), arr.NapiValue(), index, value.NapiValue()))
}

// Get Value from index
func (arr *Array) Get(index int) (ValueType, error) {
	napiValue, err := mustValueErr(napi.GetElement(arr.NapiEnv(), arr.NapiValue(), index))
	if err != nil {
		return nil, err
	}
	return N_APIValue(arr.Env(), napiValue), nil
}

// Get values with [iter.Seq]
func (arr *Array) Seq() iter.Seq[ValueType] {
	length, err := arr.Length()
	if err != nil {
		return nil
	}
	return func(yield func(ValueType) bool) {
		for index := range length {
			if value, err := arr.Get(index); err == nil {
				if !yield(value) {
					return
				}
			}
		}
	}
}

func (arr *Array) Append(values ...ValueType) error {
	length, err := arr.Length()
	if err != nil {
		return err
	}
	for valueIndex := range values {
		if err = arr.Set(length+valueIndex, values[valueIndex]); err != nil {
			return err
		}
	}
	return nil
}
