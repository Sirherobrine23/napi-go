package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

type Refence struct {
	ref        napi.Reference
	currentRef int
}

// Return new reference to Value
func CreateReference(valueOf ValueType) (*Refence, error) {
	env := valueOf.Env().NapiValue()
	ref, err := napi.CreateReference(env, valueOf.NapiValue(), 0)
	if err.ToError() != nil {
		return nil, err.ToError()
	}

	return &Refence{ref: ref, currentRef: 0}, nil
}

// Deletes the reference passed
func (ref *Refence) RemoveRef(env EnvType) error {
	return napi.DeleteReference(env.NapiValue(), ref.ref).ToError()
}

// Increments the reference count for the reference passed in
func (ref *Refence) Ref(env EnvType) error {
	newRef, err := napi.ReferenceRef(env.NapiValue(), ref.ref)
	if err := err.ToError(); err != nil {
		return err
	}
	ref.currentRef = newRef
	return nil
}

// Decrements the reference count for the reference passed in
func (ref *Refence) Unref(env EnvType) error {
	newRef, err := napi.ReferenceUnref(env.NapiValue(), ref.ref)
	if err := err.ToError(); err != nil {
		return err
	}
	ref.currentRef = newRef
	return nil
}

// Get value if is alive else return error
func (ref *Refence) ValeuOf(env EnvType) (ValueType, error) {
	valueOf, err := napi.GetReferenceValue(env.NapiValue(), ref.ref)
	if err := err.ToError(); err != nil {
		return nil, err
	}
	return N_APIValue(env, valueOf), nil
}
