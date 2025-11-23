//go:build !windows
package client

import (
	"context"
	"net"
)

func dial(ctx context.Context, addr string) (net.Conn, error) {
	network, address := parseDialTarget(addr)
	return (&net.Dialer{}).DialContext(ctx, network, address)
}
