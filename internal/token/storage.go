package token

import "github.com/zalando/go-keyring"

type tokenStorage struct{}

func (ts tokenStorage) Get(service, user string) (string, error) {
	return keyring.Get(service, user)
}

func (ts tokenStorage) Set(service, user, password string) error {
	return keyring.Set(service, user, password)
}

func (ts tokenStorage) Delete(service, user string) error {
	return keyring.Delete(service, user)
}
