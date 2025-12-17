package smbclient

import (
	"encoding/hex"
	"fmt"

	"github.com/hirochachacha/go-smb2"
)

// AuthType enum
type AuthType int

const (
	AuthNTLM AuthType = iota
)

// InitiatorFactory logic to enable different auth types
// Note: Kerberos support is currently blocked by go-smb2 sealed interface.
func GetInitiator(user, pass, domain, hash, ccachePath, realm, krbConfig string) (*smb2.NTLMInitiator, error) {
	// 1. Kerberos via ccache - NOT SUPPORTED by go-smb2 currently
	if ccachePath != "" {
		return nil, fmt.Errorf("Kerberos auth via ccache is not supported by the underlying go-smb2 library (sealed interface). Please use NTLM/Hash.")
	}

	// 2. NTLM Hash
	if hash != "" {
		hashBytes, err := hex.DecodeString(hash)
		if err != nil {
			return nil, fmt.Errorf("invalid ntlm hash format: %v", err)
		}
		return &smb2.NTLMInitiator{
			User:   user,
			Domain: domain,
			Hash:   hashBytes,
		}, nil
	}

	// 3. User/Pass
	return &smb2.NTLMInitiator{
		User:     user,
		Password: pass,
		Domain:   domain,
	}, nil
}
