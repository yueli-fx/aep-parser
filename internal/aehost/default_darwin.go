//go:build darwin

package aehost

func DefaultHost() Host {
	return NewUnavailableHost(DefaultUnavailableReason)
}
