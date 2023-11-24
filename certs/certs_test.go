package certs

import (
	"context"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("missing", func(t *testing.T) {
		fn, err := Load(ctx, "missing", "missing")
		require.Error(t, err)
		require.Nil(t, fn)
	})
	dir, err := os.MkdirTemp("", "certs")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	var (
		want tls.Certificate
		fn   func(*tls.ClientHelloInfo) (*tls.Certificate, error)
	)
	t.Run("load", func(t *testing.T) {
		want, err = New("acme.org")
		require.NoError(t, err)
		require.NotNil(t, want.PrivateKey)
		require.NotEmpty(t, want.Certificate)
		write(t, dir, want)
		fn, err = Load(ctx, filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem"))
		require.NoError(t, err)
		require.NotNil(t, fn)
		got, err := fn(nil)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, want.Certificate, got.Certificate)
		require.Equal(t, want.PrivateKey, got.PrivateKey)
	})
	t.Run("reload", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			want, err = New("acme.org")
			require.NoError(t, err)
			write(t, dir, want)
			time.Sleep(100 * time.Millisecond)
			got, err := fn(nil)
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, want.Certificate, got.Certificate)
			require.Equal(t, want.PrivateKey, got.PrivateKey)
			require.Equal(t, want.Leaf, got.Leaf)
		}
	})
	t.Run("removed", func(t *testing.T) {
		require.NoError(t, os.Remove(filepath.Join(dir, "cert.pem")))
		require.NoError(t, os.Remove(filepath.Join(dir, "key.pem")))
		got, err := fn(nil)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, want.Certificate, got.Certificate)
		require.Equal(t, want.PrivateKey, got.PrivateKey)
	})
}

func write(t *testing.T, dir string, cert tls.Certificate) {
	crt, err := os.Create(filepath.Join(dir, "cert.pem"))
	require.NoError(t, err)
	defer crt.Close()
	require.NoError(t, pem.Encode(crt, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Certificate[0],
	}))
	if err := crt.Sync(); err != nil {
		t.Fatal(err)
	}
	key, err := os.Create(filepath.Join(dir, "key.pem"))
	require.NoError(t, err)
	defer key.Close()
	b, err := x509.MarshalECPrivateKey(cert.PrivateKey.(*ecdsa.PrivateKey))
	require.NoError(t, err)
	require.NoError(t, pem.Encode(key, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: b,
	}))
	if err := key.Sync(); err != nil {
		t.Fatal(err)
	}
}
