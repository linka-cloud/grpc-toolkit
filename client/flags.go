package client

import (
	"fmt"
	"strings"

	"github.com/caitlinelfring/go-env-default"
	"github.com/spf13/pflag"
)

var u = strings.ToUpper

func NewFlagSet() (*pflag.FlagSet, Option) {
	const (
		addr   = "address"
		secure = "secure"
		// caCert     = "ca-cert"
		// clientCert = "client-cert"
		// clientKey  = "client-key"
	)
	var (
		optAddress string
		optSecure  bool
		// optCACert  string
		// optCert    string
		// optKey     string
	)
	flags := pflag.NewFlagSet("gRPC", pflag.ContinueOnError)
	flags.StringVar(&optAddress, addr, env.GetDefault(u(addr), "0.0.0.0:0"), "Bind address for the server. 127.0.0.1:9090"+flagEnv(addr))
	flags.BoolVar(&optSecure, secure, env.GetBoolDefault(u(secure), true), "Generate self signed certificate if none provided"+flagEnv(secure))
	// flags.StringVar(&optCACert, caCert, "", "Path to Root CA certificate"+flagEnv(optCACert))
	// flags.StringVar(&optCert, clientCert, "", "Path to Server certificate"+flagEnv(clientCert))
	// flags.StringVar(&optKey, clientKey, "", "Path to Server key"+flagEnv(clientKey))
	return flags, func(o *options) {
		o.addr = optAddress
		o.secure = optSecure

	}
}

func flagEnv(name string) string {
	return fmt.Sprintf(" [$%s]", strings.Replace(u(name), "-", "_", -1))
}
