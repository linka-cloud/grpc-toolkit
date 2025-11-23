package peercreds

import (
	"context"
	"errors"
	"net"

	"github.com/tailscale/peercred"
	"google.golang.org/grpc/credentials"
)

var _ credentials.TransportCredentials = (*peerCreds)(nil)

var common = credentials.CommonAuthInfo{SecurityLevel: credentials.PrivacyAndIntegrity}

func New() credentials.TransportCredentials {
	return &peerCreds{info: credentials.ProtocolInfo{
		SecurityProtocol: "peercred",
		ProtocolVersion:  "0.1",
	}}
}

type peerCreds struct {
	info credentials.ProtocolInfo
}

// AuthInfo we’ll attach to the gRPC peer
type AuthInfo struct {
	credentials.CommonAuthInfo
	Creds *peercred.Creds
}

func (AuthInfo) AuthType() string { return "peercred" }

func (t *peerCreds) ClientHandshake(_ context.Context, _ string, conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return t.handshakeConn(conn)
}

func (t *peerCreds) ServerHandshake(conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return t.handshakeConn(conn)
}

func (t *peerCreds) Info() credentials.ProtocolInfo {
	return t.info
}

func (t *peerCreds) Clone() credentials.TransportCredentials {
	return &peerCreds{info: t.info}
}

func (t *peerCreds) OverrideServerName(name string) error {
	t.info.ServerName = name
	return nil
}

func (t *peerCreds) handshakeConn(conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	if conn.RemoteAddr().Network() != "unix" {
		return nil, nil, errors.New("peercred only works with unix domain sockets")
	}
	creds, err := peercred.Get(conn)
	if err != nil {
		if errors.Is(err, peercred.ErrNotImplemented) {
			return nil, nil, errors.New("peercred not implemented on this OS")
		}
		return nil, nil, err
	}
	return conn, AuthInfo{Creds: creds, CommonAuthInfo: common}, nil
}
