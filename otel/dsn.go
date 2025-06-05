package otel

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
)

type DSN struct {
	original string

	Scheme   string
	Host     string
	HTTPPort string
	User     string
	Password string
	Token    string
}

func (dsn *DSN) String() string {
	return dsn.original
}

func (dsn *DSN) SiteURL() string {
	return dsn.Scheme + "://" + joinHostPort(dsn.Host, dsn.HTTPPort)
}

func (dsn *DSN) OTLPHttpEndpoint() string {
	return joinHostPort(dsn.Host, dsn.HTTPPort)
}

func (dsn *DSN) Headers() map[string]string {
	if dsn.Token != "" {
		return map[string]string{
			"Authorization": "Bearer " + dsn.Token,
		}
	}
	if dsn.User != "" && dsn.Password != "" {
		return map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(dsn.User+":"+dsn.Password)),
		}
	}
	return nil
}

func ParseDSN(dsnStr string) (*DSN, error) {
	if dsnStr == "" {
		return nil, fmt.Errorf("DSN is empty (use WithDSN or OTEL_DSN env var)")
	}

	u, err := url.Parse(dsnStr)
	if err != nil {
		return nil, fmt.Errorf("can't parse DSN=%q: %s", dsnStr, err)
	}

	switch u.Scheme {
	case "http", "https":
	case "":
		return nil, fmt.Errorf("DSN=%q does not have a scheme", dsnStr)
	default:
		return nil, fmt.Errorf("DSN=%q has unsupported scheme %q", dsnStr, u.Scheme)
	}

	if u.Host == "" {
		return nil, fmt.Errorf("DSN=%q does not have a host", dsnStr)
	}

	dsn := DSN{
		original: dsnStr,
		Scheme:   u.Scheme,
		Host:     u.Host,
	}
	if p, ok := u.User.Password(); ok {
		dsn.User = u.User.Username()
		dsn.Password = p
	} else {
		dsn.Token = u.User.Username()
	}

	if host, port, err := net.SplitHostPort(u.Host); err == nil {
		dsn.Host = host
		dsn.HTTPPort = port
	}

	if dsn.HTTPPort == "" {
		switch dsn.Scheme {
		case "http":
			dsn.HTTPPort = "80"
		case "https":
			dsn.HTTPPort = "443"
		}
	}

	return &dsn, nil
}

func joinHostPort(host, port string) string {
	if port == "" {
		return host
	}
	return net.JoinHostPort(host, port)
}
