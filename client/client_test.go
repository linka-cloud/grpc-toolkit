package client

import (
	"testing"
)

func TestParseDialTarget(t *testing.T) {
	tests := []struct {
		input        string
		expectedNet  string
		expectedAddr string
	}{
		{"tcp://localhost:50051", "tcp", "localhost:50051"},
		{"localhost:50051", "tcp", "localhost:50051"},
		{"unix:///tmp/socket", "unix", "/tmp/socket"},
		{"unix://C:/path/to/socket", "unix", "C:/path/to/socket"},
		{"unix:path/to/socket", "unix", "path/to/socket"},
		{`\\.\pipe\example`, "pipe", `\\.\pipe\example`},
	}
	for _, test := range tests {
		net, addr := parseDialTarget(test.input)
		if net != test.expectedNet || addr != test.expectedAddr {
			t.Errorf("parseDialTarget(%q) = (%q, %q); want (%q, %q)", test.input, net, addr, test.expectedNet, test.expectedAddr)
		}
	}
}
