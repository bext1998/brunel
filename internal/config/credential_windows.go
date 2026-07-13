//go:build windows

package config

import (
	"context"
	"errors"
	"syscall"
	"unsafe"
)

const credentialTypeGeneric = 1

type credential struct {
	Flags              uint32
	Type               uint32
	TargetName         *uint16
	Comment            *uint16
	LastWritten        syscall.Filetime
	CredentialBlobSize uint32
	CredentialBlob     *byte
	Persist            uint32
	AttributeCount     uint32
	Attributes         uintptr
	TargetAlias        *uint16
	UserName           *uint16
}

var (
	advapi32 = syscall.NewLazyDLL("advapi32.dll")
	credRead = advapi32.NewProc("CredReadW")
	credFree = advapi32.NewProc("CredFree")
)

type platformCredentialSource struct{}

func NewPlatformCredentialSource() CredentialSource { return platformCredentialSource{} }

func (platformCredentialSource) OpenRouterAPIKey(context.Context) (string, error) {
	target, err := syscall.UTF16PtrFromString(OpenRouterCredentialTarget)
	if err != nil {
		return "", err
	}
	var pointer *credential
	result, _, callErr := credRead.Call(
		uintptr(unsafe.Pointer(target)),
		credentialTypeGeneric,
		0,
		uintptr(unsafe.Pointer(&pointer)),
	)
	if result == 0 {
		if callErr == nil {
			return "", errors.New("Credential Manager read failed")
		}
		return "", callErr
	}
	defer credFree.Call(uintptr(unsafe.Pointer(pointer)))
	if pointer == nil || pointer.CredentialBlob == nil || pointer.CredentialBlobSize == 0 {
		return "", errors.New("Credential Manager credential is empty")
	}
	return string(unsafe.Slice(pointer.CredentialBlob, pointer.CredentialBlobSize)), nil
}
