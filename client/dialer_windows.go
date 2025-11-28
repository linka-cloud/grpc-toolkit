//go:build windows

package client

import (
	"context"
	"net"

	"golang.zx2c4.com/wireguard/ipc/namedpipe"
)

func dial(ctx context.Context, addr string) (net.Conn, error) {
	network, address := parseDialTarget(addr)
	if network == "pipe" {
		return namedpipe.DialContext(ctx, address)
	}
	return (&net.Dialer{}).DialContext(ctx, network, address)
}
