package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

// Process status to return error if StatusOK return nil on error
func mustValueErr[T any](input T, status napi.Status) (T, error) {
	if status != napi.StatusOK {
		return input, napi.StatusError(status)
	}
	return input, nil
}

// return error from status
func singleMustValueErr(status napi.Status) error {
	if status != napi.StatusOK {
		return napi.StatusError(status)
	}
	return nil
}

// Process status to return error if StatusOK return nil on error
func mustValueErr2[T any](input T, _ bool, status napi.Status) (T, error) {
	if status != napi.StatusOK {
		return input, napi.StatusError(status)
	}
	return input, nil
}

// Process status to return error if StatusOK return nil on error
func mustValueErr3[T, C any](input T, i2 C, status napi.Status) (T, C, error) {
	if status != napi.StatusOK {
		return input, i2, napi.StatusError(status)
	}
	return input, i2, nil
}
