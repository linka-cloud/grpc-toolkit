package service

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cmd = &cobra.Command{
	Short: "A gRPC micro service",
}

const (
	caCert = "ca_cert"
	serverAddress = "server_address"
	serverCert = "server_cert"
	serverKey = "server_key"
)

func init() {
	viper.AutomaticEnv()
	// server_address
	cmd.Flags().String(serverAddress, "0.0.0.0:0", "Bind address for the server. 127.0.0.1:9090 [$SERVER_ADDRESS]")
	viper.BindPFlag(serverAddress, cmd.Flags().Lookup(serverAddress))
	// ca_cert
	cmd.Flags().String(caCert, "", "Path to Root CA certificate [$CA_CERT]")
	viper.BindPFlag(caCert, cmd.Flags().Lookup(caCert))
	// server_cert
	cmd.Flags().String(serverCert, "", "Path to Server certificate [$SERVER_CERT]")
	viper.BindPFlag(serverCert, cmd.Flags().Lookup(serverCert))
	// server_key
	cmd.Flags().String(serverKey, "", "Path to Server key [$SERVER_KEY]")
	viper.BindPFlag(serverKey, cmd.Flags().Lookup(serverKey))
}

func parseFlags(o *options) *options {
	o.address = viper.GetString(serverAddress)
	o.caCert = viper.GetString(caCert)
	o.cert = viper.GetString(serverCert)
	o.key = viper.GetString(serverKey)
	return o
}