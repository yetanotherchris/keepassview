//go:build windows

package creds

import (
	"fmt"

	"github.com/danieljoos/wincred"
)

// New returns the Windows credential store implementation.
func New() Store { return &wincredStore{} }

type wincredStore struct{}

func (*wincredStore) Available() bool { return true }

func (*wincredStore) Get(key string) (string, error) {
	cred, err := wincred.GetGenericCredential(key)
	if err != nil {
		return "", fmt.Errorf("get credential %q: %w", key, err)
	}
	return string(cred.CredentialBlob), nil
}

func (*wincredStore) Set(key, value string) error {
	cred := wincred.NewGenericCredential(key)
	cred.CredentialBlob = []byte(value)
	if err := cred.Write(); err != nil {
		return fmt.Errorf("set credential %q: %w", key, err)
	}
	return nil
}

func (*wincredStore) Delete(key string) error {
	cred, err := wincred.GetGenericCredential(key)
	if err != nil {
		return nil // already gone
	}
	return cred.Delete()
}
