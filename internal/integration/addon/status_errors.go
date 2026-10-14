package main

import (
	"errors"
	"fmt"
	"strings"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	internalnapi "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

func checkStatusErrors() error {
	if internalnapi.StatusOK.ToError() != nil {
		return fmt.Errorf("napi_ok must not return an error")
	}
	cases := []struct {
		status   internalnapi.Status
		sentinel error
	}{
		{internalnapi.StatusClosing, napi.ErrThreadsafeFunctionClosing},
		{internalnapi.StatusQueueFull, napi.ErrThreadsafeFunctionQueueFull},
	}
	for _, test := range cases {
		actual := test.status.ToError()
		wrapped := fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", actual))
		if !errors.Is(actual, test.sentinel) || !errors.Is(wrapped, test.sentinel) {
			return fmt.Errorf("%s did not match its public sentinel", test.status)
		}
		if !errors.Is(wrapped, test.status.ToError()) {
			return fmt.Errorf("independent stack traces changed %s identity", test.status)
		}
		pointerTarget := &internalnapi.StatusError{Status: test.status}
		if !errors.Is(wrapped, pointerTarget) || !errors.Is(pointerTarget, test.sentinel) {
			return fmt.Errorf("pointer/value status targets did not match")
		}
		var nilTarget *internalnapi.StatusError
		if errors.Is(wrapped, nilTarget) || errors.Is(wrapped, nil) {
			return fmt.Errorf("nil targets must not match")
		}
		if errors.Is(wrapped, internalnapi.StatusInvalidArg.ToError()) || errors.Is(wrapped, errors.New(actual.Error())) {
			return fmt.Errorf("unrelated errors must not match %s", test.status)
		}
		var recovered internalnapi.StatusError
		if !errors.As(wrapped, &recovered) || recovered.Status != test.status {
			return fmt.Errorf("errors.As lost the native status code")
		}
		if !strings.Contains(recovered.Error(), "goroutine ") {
			return fmt.Errorf("native error stack trace was lost")
		}
	}
	if errors.Is(napi.ErrThreadsafeFunctionClosing, napi.ErrThreadsafeFunctionQueueFull) {
		return fmt.Errorf("closing and queue-full must remain distinct")
	}
	return nil
}

func registerStatusErrors(env napi.EnvType, export *napi.Object) {
	callback := must(napi.CreateFunction(env, "statusErrors", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		return nil, checkStatusErrors()
	}))
	if err := export.Set("statusErrors", callback); err != nil {
		panic(err)
	}
}
