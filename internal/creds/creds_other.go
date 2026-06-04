//go:build !windows

package creds

import "errors"

// New returns a stub that reports the credential store as unavailable.
func New() Store { return &stubStore{} }

type stubStore struct{}

func (*stubStore) Available() bool                    { return false }
func (*stubStore) Get(string) (string, error)         { return "", errors.New("credential store not supported on this platform") }
func (*stubStore) Set(string, string) error           { return errors.New("credential store not supported on this platform") }
func (*stubStore) Delete(string) error                { return nil }
