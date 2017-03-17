package whatip

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenDNS(t *testing.T) {
	t.Parallel()
	ip, err := OpenDNS.GetIP()
	require.NoError(t, err)
	assert.NotNil(t, ip)
}
