package service

import (
	"net"

	"github.com/Microsoft/go-winio"
)

func listen(network, address string) (net.Listener, error) {
	if network == "pipe" {
		return winio.ListenPipe(address, nil)
	}
	return net.Listen(network, address)
}
