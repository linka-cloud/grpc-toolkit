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
	"strings"
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

func TLSConfig(ctx context.Context, cert, key string) (*tls.Config, error) {
	c, err := Load(ctx, cert, key)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		GetCertificate: c,
	}, nil
}

func Load(ctx context.Context, cert, key string) (func(info *tls.ClientHelloInfo) (*tls.Certificate, error), error) {
	c, err := file.NewConfig(cert)
	if err != nil {
		return nil, fmt.Errorf("failed to load cert: %v", err)
	}
	k, err := file.NewConfig(key)
	if err != nil {
		return nil, fmt.Errorf("failed to load key: %v", err)
	}
	crt, err := load(c, key)
	if err != nil {
		return nil, fmt.Errorf("failed to load cert: %v", err)
	}
	var mu sync.RWMutex
	kch := make(chan []byte)
	if err := k.Watch(ctx, kch); err != nil {
		return nil, fmt.Errorf("failed to watch key: %v", err)
	}
	cch := make(chan []byte)
	if err := c.Watch(ctx, cch); err != nil {
		return nil, fmt.Errorf("failed to watch cert: %v", err)
	}
	reload := func() {
		c, err := load(c, key)
		// ignore errors due to cert and key not matching as this is expected
		// when the cert is being reloaded and the key is not yet updated or vice versa
		if err != nil && !strings.Contains(err.Error(), "does not match") {
			logger.C(ctx).Errorf("failed to reload cert: %v", err)
			return
		}
		mu.Lock()
		crt = c
		mu.Unlock()
	}
	go func() {
		for {
			select {
			case <-kch:
				reload()
			case <-cch:
				reload()
			case <-ctx.Done():
				return
			}
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
