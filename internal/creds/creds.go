// Package creds wraps the OS credential store behind a small interface.
package creds

// Store provides get/set/delete access to the OS credential store.
type Store interface {
	Available() bool
	Get(key string) (string, error)
	Set(key, value string) error
	Delete(key string) error
}
