package peercreds

import (
	"context"
	"crypto/tls"
	"errors"
	"net"

	"github.com/soheilhy/cmux"
	"github.com/tailscale/peercred"
	"google.golang.org/grpc/credentials"
)

var ErrUnsupportedConnType = peercred.ErrUnsupportedConnType

var _ credentials.TransportCredentials = (*peerCreds)(nil)

// Creds are the peer credentials.
type Creds struct {
	pid int
	uid string
}

func (c *Creds) PID() (pid int, ok bool) {
	return c.pid, c.pid != 0
}

// UserID returns the userid (or Windows SID) that owns the other side
// of the connection, if known. (ok is false if not known)
// The returned string is suitable to passing to os/user.LookupId.
func (c *Creds) UserID() (uid string, ok bool) {
	return c.uid, c.uid != ""
}

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
	Creds Creds
}

func (AuthInfo) AuthType() string { return "peercred" }

func (t *peerCreds) ClientHandshake(ctx context.Context, authority string, conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
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
	if conn.RemoteAddr().Network() != "unix" && conn.RemoteAddr().Network() != "pipe" {
		return nil, nil, errors.New("peercred only works with unix domain sockets or Windows named pipes")
	}
	inner := conn
unwrap:
	for {
		switch c := inner.(type) {
		case *cmux.MuxConn:
			inner = c.Conn
		case *tls.Conn:
			inner = c.NetConn()
		default:
			break unwrap
		}
	}
	creds, err := Get(inner)
	if err != nil {
		if errors.Is(err, peercred.ErrNotImplemented) {
			return nil, nil, errors.New("peercred not implemented on this OS")
		}
		return nil, nil, err
	}
	var c Creds
	c.uid, _ = creds.UserID()
	c.pid, _ = creds.PID()
	return conn, AuthInfo{Creds: c, CommonAuthInfo: common}, nil
}
