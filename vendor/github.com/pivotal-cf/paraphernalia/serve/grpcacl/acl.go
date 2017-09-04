// Package grpcacl provides a method of applying coarse service ACLs based on
// the identity of the connecting service.
//
// The certificate format expected by this package *is not* stable. Please do
// not use it in production.
package grpcacl

import (
	"crypto/tls"
	"errors"
	"net"

	"google.golang.org/grpc/credentials"
)

// ErrClientNotApproved is the error returned when a client that is not allowed
// attempts to connect to the server.
var ErrClientNotApproved = errors.New("client not an allowed identity")

// ErrNonTLSTransport is returned if this credential is used on a transport
// without TLS.
var ErrNonTLSTransport = errors.New("transport is not tls")

type acl struct {
	credentials.TransportCredentials

	allowed map[string]struct{}

	origConfig  *tls.Config
	origAllowed []string
}

func (a *acl) ServerHandshake(rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	conn, info, err := a.TransportCredentials.ServerHandshake(rawConn)
	if err != nil {
		return nil, nil, err
	}

	tlsInfo, ok := info.(credentials.TLSInfo)
	if !ok {
		return nil, info, ErrNonTLSTransport
	}

	// XXX: Is it possible for us to get this point with no certificates?
	cn := tlsInfo.State.PeerCertificates[0].Subject.CommonName
	_, found := a.allowed[cn]
	if !found {
		return nil, nil, ErrClientNotApproved
	}

	return conn, info, nil
}

func (a *acl) Clone() credentials.TransportCredentials {
	return NewTLS(a.origConfig, a.origAllowed...)
}

// NewTLS creates a new transport credential that verifies that any connecting
// client has a common name from the allowed list.
func NewTLS(c *tls.Config, allowed ...string) credentials.TransportCredentials {
	creds := credentials.NewTLS(c)

	allow := make(map[string]struct{})
	for _, a := range allowed {
		allow[a] = struct{}{}
	}

	return &acl{
		TransportCredentials: creds,
		allowed:              allow,

		origConfig:  c,
		origAllowed: allowed,
	}
}
