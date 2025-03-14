package log_test

import (
	"github.com/ekazakas/skibidi/pkg/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNoopAdapter_WithEquality(t *testing.T) {
	adapter := log.NoopAdapter{}

	assert.Equal(t, adapter, adapter.With(log.Fields{}))
}
