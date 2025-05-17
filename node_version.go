package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

type NodeVersion = napi.NodeVersion

// GetNodeVersion retrieves the Node.js version information for the given environment.
// It takes an EnvType as input and returns a pointer to a NodeVersion struct and an error.
// If the version retrieval fails, it returns a non-nil error.
func GetNodeVersion(env EnvType) (*NodeVersion, error) {
	version, err := napi.GetNodeVersion(env.NapiValue())
	if err := err.ToError(); err != nil {
		return nil, err
	}
	return &version, nil
}