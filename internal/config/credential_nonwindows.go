//go:build !windows

package config

import (
	"context"
)

type platformCredentialSource struct{}

func NewPlatformCredentialSource() CredentialSource { return platformCredentialSource{} }

func (platformCredentialSource) OpenRouterAPIKey(context.Context) (string, error) {
	return "", ErrUnsupportedPlatform
}
