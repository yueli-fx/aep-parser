//go:build !windows && !darwin

package aehost

func DefaultHost() Host {
	return NewUnavailableHost(DefaultUnavailableReason)
}
