package service

import (
	"fmt"
	"strings"

	"github.com/caitlinelfring/go-env-default"
	"github.com/spf13/pflag"
)

const (
	serverAddress = "address"

	insecure   = "insecure"
	reflection = "reflection"

	caCert     = "ca-cert"
	serverCert = "server-cert"
	serverKey  = "server-key"

	clientCACert = "client-ca-cert"
	clientCert   = "client-cert"
	clientKey    = "client-key"
)

var u = strings.ToUpper

func NewFlagSet() (*pflag.FlagSet, Option) {
	var (
		optAddress    string
		optInsecure   bool
		optReflection bool
		optCACert     string
		optCert       string
		optKey        string
	)
	flags := pflag.NewFlagSet("gRPC", pflag.ContinueOnError)
	flags.StringVarP(&optAddress, serverAddress, "a", env.GetDefault(u(serverAddress), "0.0.0.0:0"), "Bind address for the server, e.g. 127.0.0.1:9090"+flagEnv(serverAddress))
	flags.BoolVar(&optInsecure, insecure, env.GetBoolDefault(u(insecure), false), "Do not generate self signed certificate if none provided"+flagEnv(insecure))
	flags.BoolVar(&optReflection, reflection, env.GetBoolDefault(u(reflection), false), "Enable gRPC reflection server"+flagEnv(reflection))
	flags.StringVar(&optCACert, caCert, "", "Path to Root CA certificate"+flagEnv(caCert))
	flags.StringVar(&optCert, serverCert, "", "Path to Server certificate"+flagEnv(serverCert))
	flags.StringVar(&optKey, serverKey, "", "Path to Server key"+flagEnv(serverKey))
	flags.StringVar(&optCACert, clientCACert, "", "Path to Root CA certificate"+flagEnv(clientCACert))
	flags.StringVar(&optCert, clientCert, "", "Path to Client certificate"+flagEnv(clientCert))
	flags.StringVar(&optKey, clientKey, "", "Path to Client key"+flagEnv(clientKey))
	return flags, func(o *options) {
		o.address = optAddress
		o.secure = !optInsecure
		o.reflection = optReflection
		o.caCert = optCACert
		o.cert = optCert
		o.key = optKey
		o.clientCACert = optCACert
		o.clientCert = optCert
		o.clientKey = optKey
	}
}

func flagEnv(name string) string {
	return fmt.Sprintf(" [$%s]", strings.Replace(u(name), "-", "_", -1))
}
