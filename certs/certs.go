package certs

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"sync"
	"time"

	"go.linka.cloud/grpc-toolkit/config"
	"go.linka.cloud/grpc-toolkit/config/file"
	"go.linka.cloud/grpc-toolkit/logger"
)

func New(host ...string) (tls.Certificate, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(time.Hour * 24 * 365)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range host {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageCertSign

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	// create public key
	certOut := bytes.NewBuffer(nil)
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	// create private key
	keyOut := bytes.NewBuffer(nil)
	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, err
	}
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})

	return tls.X509KeyPair(certOut.Bytes(), keyOut.Bytes())
}

func Load(ctx context.Context, cert, key string) (func(info *tls.ClientHelloInfo) (*tls.Certificate, error), error) {
	f, err := file.NewConfig(cert)
	if err != nil {
		return nil, fmt.Errorf("failed to load cert: %v", err)
	}
	crt, err := load(f, key)
	if err != nil {
		return nil, fmt.Errorf("failed to load cert: %v", err)
	}
	var mu sync.RWMutex
	ch := make(chan []byte)
	if err := f.Watch(ctx, ch); err != nil {
		return nil, fmt.Errorf("failed to watch cert: %v", err)
	}
	go func() {
		for range ch {
			c, err := load(f, key)
			if err != nil {
				logger.C(ctx).Errorf("failed to reload cert: %v", err)
				continue
			}
			mu.Lock()
			crt = c
			mu.Unlock()
		}
	}()

	return func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
		mu.RLock()
		defer mu.RUnlock()
		return crt, nil
	}, nil
}

func load(cert config.Config, key string) (*tls.Certificate, error) {
	cb, err := cert.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read cert: %v", err)
	}
	kb, err := os.ReadFile(key)
	if err != nil {
		return nil, fmt.Errorf("failed to read key: %v", err)
	}
	c, err := tls.X509KeyPair(cb, kb)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
