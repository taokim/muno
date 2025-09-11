package navigator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFactoryHelpers(t *testing.T) {
	t.Run("NewFilesystem", func(t *testing.T) {
		workspace := t.TempDir()
		nav, err := NewFilesystem(workspace, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, nav)
		_, ok := nav.(*FilesystemNavigator)
		assert.True(t, ok)
	})

	t.Run("NewInMemory", func(t *testing.T) {
		nav := NewInMemory()
		assert.NotNil(t, nav)
		_, ok := nav.(*InMemoryNavigator)
		assert.True(t, ok)
	})

	t.Run("NewCached", func(t *testing.T) {
		workspace := t.TempDir()
		nav, err := NewCached(workspace, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, nav)
		_, ok := nav.(*CachedNavigator)
		assert.True(t, ok)
	})

	t.Run("WithMaxSize", func(t *testing.T) {
		base := NewInMemory()
		cached := NewCachedNavigator(base, 100*time.Millisecond)
		result := cached.WithMaxSize(50)
		assert.NotNil(t, result)
		// Verify it returns the same instance (fluent API)
		assert.Equal(t, cached, result)
	})
}