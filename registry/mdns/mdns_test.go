package mdns

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.bertha.cloud/partitio/lab/grpc/registry"
)

func TestRegistry(t *testing.T) {
	assert := assert.New(t)
	reg := NewRegistry()
	svc := &registry.Service{Name: "test", Nodes: []*registry.Node{{Id: "test-1", Address: "127.0.0.1:8888"}}}
	if err := reg.Register(svc); err != nil {
		t.Fatal(err)
	}
	svcs, err := reg.GetService("test")
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(svcs, 1)
	assert.Contains(svcs, svc)
	if err := reg.Deregister(svc); err != nil {
		t.Fatal(err)
	}
	svcs, err = reg.ListServices()
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(svcs, 0)
}
