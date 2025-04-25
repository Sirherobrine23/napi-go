package napi

/*
#cgo CFLAGS: -DDEBUG
#cgo CFLAGS: -D_DEBUG
#cgo CFLAGS: -DV8_ENABLE_CHECKS
#cgo CFLAGS: -DNAPI_EXPERIMENTAL
#cgo CFLAGS: -I/usr/local/include/node
#cgo CXXFLAGS: -std=c++11

#cgo darwin LDFLAGS: -Wl,-undefined,dynamic_lookup
#cgo darwin LDFLAGS: -Wl,-no_pie
#cgo darwin LDFLAGS: -Wl,-search_paths_first
// #cgo darwin amd64 LDFLAGS: -arch x86_64
// #cgo darwin arm64 LDFLAGS: -arch arm64

#cgo linux LDFLAGS: -Wl,-unresolved-symbols=ignore-all

#cgo LDFLAGS: -L${SRCDIR}
*/
import "C"

// Process status to return error if StatusOK return nil on error
func MustValueErr2[T any](input T, _ bool, status Status) (T, error) {
	if status != StatusOK {
		return input, StatusError(status)
	}
	return input, nil
}

// Process status to return error if StatusOK return nil on error
func MustValueErr3[T, C any](input T, i2 C, status Status) (T, C, error) {
	if status != StatusOK {
		return input, i2, StatusError(status)
	}
	return input, i2, nil
}

// Process status to return error if StatusOK return nil on error
func MustValueErr[T any](input T, status Status) (T, error) {
	if status != StatusOK {
		return input, StatusError(status)
	}
	return input, nil
}

func SingleMustValueErr(status Status) error {
	if status != StatusOK {
		return StatusError(status)
	}
	return nil
}

// Create panic if status return Error
func MustValue[T any](input T, status Status) T {
	v, err := MustValueErr(input, status)
	if err != nil {
		panic(err)
	}
	return v
}
