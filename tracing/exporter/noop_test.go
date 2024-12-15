package exporter_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/lib-go/tracing/exporter"
)

func TestNoopExporter(t *testing.T) {
	noopExporter := exporter.NewNoopExporter()

	t.Parallel()

	t.Run("NoopExporter", func(t *testing.T) {
		t.Parallel()
		assert.NotNil(t, noopExporter)
	})

	t.Run("ExportSpans", func(t *testing.T) {
		t.Parallel()
		err := noopExporter.ExportSpans(context.Background(), nil)
		assert.Nil(t, err)
	})

	t.Run("Shutdown", func(t *testing.T) {
		t.Parallel()
		err := noopExporter.Shutdown(context.Background())
		assert.Nil(t, err)
	})
}
