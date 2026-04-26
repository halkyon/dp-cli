package filters

import (
	"testing"

	"github.com/halkyon/dp/testapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAliases_Get(t *testing.T) {
	var mq testapi.MockQuerier

	aliases := NewAliases(&mq, 0)

	result, err := aliases.Get(t.Context())
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"test-server-1", "test-server-2"}, result)
}
