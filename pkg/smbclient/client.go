package smbclient

import (
	"fmt"
	"net"

	"github.com/hirochachacha/go-smb2"
)

type Session struct {
	Session *smb2.Session
	Conn    net.Conn
}

func NewSession(host, username, password, domain, hash, ccache, krbConfig string) (s *Session, err error) {
	conn, err := net.Dial("tcp", host+":445")
	if err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in SMB dial: %v", r)
			if conn != nil {
				conn.Close()
			}
		}
	}()

	initiator, err := GetInitiator(username, password, domain, hash, ccache, "", krbConfig)
	if err != nil {
		conn.Close()
		return nil, err
	}

	d := &smb2.Dialer{
		Initiator: initiator,
		// Enforce signing if possible, which might help match server expectations
		// or at least standardizes behavior against DCs.
		// However, Negotiator definition in go-smb2 might be an interface or struct.
		// Checking standard usage: usually one doesn't set it unless needed.
		// Let's try explicit DefaultNegotiator equivalent but with signing?
		// go-smb2 doesn't export a simple struct with bools easily in all versions.
		// Safest bet is to leave it default but handle the error.
		// Wait, user says "failed auth". But ListShares failed. Authentication passed.
	}

	// TODO: Verify Hash support in go-smb2 or use ntlmssp directly if needed.
	// For now assuming User/Pass.

	session, err := d.Dial(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Session{Session: session, Conn: conn}, nil
}

func (s *Session) ListShares() ([]string, error) {
	return s.Session.ListSharenames()
}

func (s *Session) Mount(share string) (*smb2.Share, error) {
	return s.Session.Mount(share)
}

func (s *Session) Close() {
	if s.Session != nil {
		s.Session.Logoff()
	}
	if s.Conn != nil {
		s.Conn.Close()
	}
}
