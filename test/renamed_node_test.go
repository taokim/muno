package test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taokim/muno/internal/config"
)

func TestRenamedNodeEagerLoading(t *testing.T) {
	// Load the test config
	cfg, err := config.LoadTree("/tmp/test-renamed-ws/muno.yaml")
	require.NoError(t, err)
	
	t.Log("Testing eager pattern checking with renamed nodes:")
	t.Log("===================================================")
	
	for _, node := range cfg.Nodes {
		isLazy := node.IsLazy()
		status := "EAGER"
		if isLazy {
			status = "LAZY"
		}
		
		t.Logf("\nNode: %s", node.Name)
		t.Logf("  URL: %s", node.URL)
		t.Logf("  Fetch Mode: %s", node.Fetch)
		t.Logf("  Result: %s", status)
		
		if node.Name == "fulfillment" && node.URL == "https://github.com/org/fulfillment-munorepo.git" {
			// This is the key test: renamed node should still be eager based on URL
			assert.False(t, isLazy, "Node with -munorepo URL should be EAGER even if renamed")
			
			if isLazy {
				t.Error("❌ FAILED: Should be EAGER because URL contains -munorepo")
			} else {
				t.Log("✅ SUCCESS: Correctly identified as EAGER based on URL pattern")
			}
		}
		
		if node.Name == "payment-service" {
			// Regular service should be lazy
			assert.True(t, isLazy, "payment-service should be LAZY")
		}
	}
}