package service_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.linka.cloud/grpc-toolkit/service"
)

func TestServiceStop(t *testing.T) {
	t.Run("service not started", func(t *testing.T) {
		s, _ := newService(t)
		require.NotNil(t, s)
		require.NoError(t, s.Stop())
	})
	t.Run("service started", func(t *testing.T) {
		s, ready := newService(t)
		go s.Start()
		<-ready
		require.NoError(t, s.Stop())
	})
	t.Run("service already started", func(t *testing.T) {
		s, ready := newService(t)
		go s.Start()
		<-ready
		require.Error(t, s.Start())
		require.NoError(t, s.Stop())
	})
	t.Run("service already stopped", func(t *testing.T) {
		s, ready := newService(t)
		go s.Start()
		<-ready
		require.NoError(t, s.Stop())
		require.NoError(t, s.Stop())
	})
	t.Run("service started and stopped multiple times", func(t *testing.T) {
		s, ready := newService(t)
		go s.Start()
		<-ready
		require.NoError(t, s.Stop())
		go s.Start()
		<-ready
		require.NoError(t, s.Stop())
	})
}

func newService(t *testing.T) (service.Service, chan struct{}) {
	t.Helper()
	ch := make(chan struct{})
	s, err := service.New(
		service.WithHealth(true),
		service.WithAfterStart(func() error {
			ch <- struct{}{}
			return nil
		}),
	)
	require.NoError(t, err)
	return s, ch
}
