package inproc

import (
	"errors"
	"net"
	"sync"

	"github.com/fullstorydev/grpchan/inprocgrpc"

	"go.linka.cloud/grpc/transport"
)

var (
	_ transport.Transport = &InProc{}
)

func New() transport.Transport {
	return &InProc{stop: make(chan struct{})}
}

type InProc struct {
	inprocgrpc.Channel
	stop    chan struct{}
	running bool
	mu      sync.RWMutex
}

func (i *InProc) Serve(_ net.Listener) error {
	i.mu.RLock()
	running := i.running
	i.mu.RUnlock()
	if running {
		return errors.New("already running")
	}
	i.mu.Lock()
	i.running = true
	i.mu.Unlock()
	<-i.stop
	return nil
}

func (i *InProc) Stop() {
	i.mu.RLock()
	running := i.running
	i.mu.RUnlock()
	if running {
		i.stop <- struct{}{}
	}
}

func (i *InProc) GracefulStop() {
	i.Stop()
}
