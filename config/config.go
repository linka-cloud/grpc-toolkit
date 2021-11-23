package config

import (
	"context"
)

type Config interface {
	Read() ([]byte, error)
	Watch(ctx context.Context, updates chan<- []byte) error
}
