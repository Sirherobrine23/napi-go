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

// Populates the Array with elements from the provided iterator sequence.
// For each element in the sequence, it converts the value to ValueType if necessary,
// and appends it to the end of the Array. If an error occurs during conversion or insertion,
// the operation stops and the error is returned.
// Returns error if an error if any occurs during value conversion or insertion; otherwise, nil.
func (arr *Array) From(from iter.Seq[any]) (err error) {
	var currentLength int
	for value := range from {
		// Get NAPI value
		var valueOf ValueType
		switch v := value.(type) {
		case ValueType:
			valueOf = v
		default:
			if valueOf, err = ValueOf(arr.Env(), v); err != nil {
				return
			}
		}

		// Get value to last element
		if currentLength, err = arr.Length(); err != nil {
			break
		} else if err = arr.Set(currentLength, valueOf); err != nil {
			break
		}
	}
	return
}

// Append adds one or more values to the end of the array.
// It accepts a variadic number of ValueType arguments and appends each to the array,
// starting from the current length. If an error occurs during the append operation,
// it returns the error; otherwise, it returns nil.
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
