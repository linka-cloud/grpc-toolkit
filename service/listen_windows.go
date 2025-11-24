package service

import (
	"net"

	"golang.zx2c4.com/wireguard/ipc/namedpipe"
)

// listen uses wireguard's namedpipe package to listen on named pipes on Windows until
// https://github.com/golang/go/issues/49650 is resolved.
// For other networks, it falls back to the standard net.Listen.
func listen(network, address string) (net.Listener, error) {
	if network == "pipe" {
		return namedpipe.Listen(address)
	}
	return net.Listen(network, address)
}
