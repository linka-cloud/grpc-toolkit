package otel

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIDGenerator(t *testing.T) {
	ctx := context.Background()
	gen := newIDGenerator()

	traceID1, spanID1 := gen.NewIDs(ctx)
	traceID2, spanID2 := gen.NewIDs(ctx)
	require.NotEqual(t, traceID1, traceID2)
	require.NotEqual(t, spanID1, spanID2)

	spanID3 := gen.NewSpanID(ctx, traceID1)
	require.NotEqual(t, spanID1, spanID3)
}

func TestSpanPrecision(t *testing.T) {
	dur := time.Duration(math.MaxUint32) * time.Duration(spanIDPrec)
	require.Equal(t, "1193h2m47.295s", dur.String())
}
