//go:build !windows

package peercreds

import (
	"net"

	"github.com/tailscale/peercred"
)

func Get(conn net.Conn) (*Creds, error) {
	creds, err := peercred.Get(conn)
	if err != nil {
		return nil, err
	}
	var c Creds
	c.uid, _ = creds.UserID()
	c.pid, _ = creds.PID()
	return &c, nil
}
