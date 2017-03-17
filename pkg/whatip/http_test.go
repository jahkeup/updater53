package whatip

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTP(t *testing.T) {
	providers := map[string]IPer{
		"ifconfig.me":   IfconfigMeHTTP,
		"icanhazip.com": ICanHazIPHTTP,
	}

	for prov, http := range providers {
		t.Run(prov, func(t *testing.T) {
			t.Parallel()
			ip, err := http.GetIP()
			require.NoError(t, err)
			assert.NotNil(t, ip)
		})
	}
}
