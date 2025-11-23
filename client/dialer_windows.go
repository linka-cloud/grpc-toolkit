//go:build windows

package client

import (
	"context"
	"net"

	"github.com/Microsoft/go-winio"
)

func dial(ctx context.Context, addr string) (net.Conn, error) {
	network, address := parseDialTarget(addr)
	if network == "pipe" {
		return winio.DialPipeContext(ctx, address)
	}
	return (&net.Dialer{}).DialContext(ctx, network, address)
}
