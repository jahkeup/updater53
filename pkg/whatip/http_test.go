package whatip_test

import (
	"testing"

	"github.com/jahkeup/updater53/pkg/whatip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTP(t *testing.T) {
	providers := map[string]whatip.IPResolver{
		"ifconfig.me":   whatip.IfconfigMeHTTP,
		"icanhazip.com": whatip.ICanHazIPHTTP,
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
