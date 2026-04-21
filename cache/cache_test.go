package cache

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_GetSet(t *testing.T) {
	// Since New is hardcoded to use HOME, we'll use it and clean up.
	c, err := New[string]("test", time.Hour, t.TempDir())
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Remove(c.Path))
	}()

	t.Run("Set and Get", func(t *testing.T) {
		data := "hello world"
		err := c.Set(data, time.Hour)
		require.NoError(t, err)

		var got string
		ok := c.Get(&got)
		assert.True(t, ok)
		assert.Equal(t, data, got)
	})

	t.Run("Get expired", func(t *testing.T) {
		c.MaxAge = -time.Hour
		data := "expired"
		err := c.Set(data, time.Hour)
		require.NoError(t, err)

		var got string
		ok := c.Get(&got)
		assert.False(t, ok)
	})

	t.Run("Get non-existent", func(t *testing.T) {
		var got string
		ok := c.Get(&got)
		assert.False(t, ok)
	})
}
