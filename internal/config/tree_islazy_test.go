package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeDefinitionIsLazy_URLCheck(t *testing.T) {
	tests := []struct {
		name     string
		node     NodeDefinition
		expected bool
		reason   string
	}{
		{
			name: "node name differs from repo URL - should check URL",
			node: NodeDefinition{
				Name:  "fulfillment", // Renamed node, doesn't match pattern
				URL:   "https://github.com/org/fulfillment-munorepo.git",
				Fetch: FetchAuto,
			},
			expected: false, // Should be eager because URL matches pattern
			reason:   "URL contains -munorepo suffix, should be eager",
		},
		{
			name: "both node name and URL match pattern",
			node: NodeDefinition{
				Name:  "backend-meta",
				URL:   "https://github.com/org/backend-meta.git",
				Fetch: FetchAuto,
			},
			expected: false, // Should be eager
			reason:   "Both name and URL match meta pattern",
		},
		{
			name: "node name matches but URL doesn't",
			node: NodeDefinition{
				Name:  "payment-monorepo",
				URL:   "https://github.com/org/payment-service.git",
				Fetch: FetchAuto,
			},
			expected: false, // Should be eager because name matches
			reason:   "Node name matches pattern even though URL doesn't",
		},
		{
			name: "neither name nor URL matches pattern",
			node: NodeDefinition{
				Name:  "payment",
				URL:   "https://github.com/org/payment-service.git",
				Fetch: FetchAuto,
			},
			expected: true, // Should be lazy
			reason:   "Neither name nor URL matches any eager pattern",
		},
		{
			name: "URL with SSH format",
			node: NodeDefinition{
				Name:  "fulfillment",
				URL:   "git@github.com:org/fulfillment-munorepo.git",
				Fetch: FetchAuto,
			},
			expected: false, // Should be eager because URL matches pattern
			reason:   "SSH URL format should also be parsed correctly",
		},
		{
			name: "URL without .git suffix",
			node: NodeDefinition{
				Name:  "fulfillment",
				URL:   "https://github.com/org/fulfillment-munorepo",
				Fetch: FetchAuto,
			},
			expected: false, // Should be eager
			reason:   "URL without .git suffix should still match pattern",
		},
		{
			name: "empty URL should only check name",
			node: NodeDefinition{
				Name:  "backend-meta",
				URL:   "",
				Fetch: FetchAuto,
			},
			expected: false, // Should be eager because name matches
			reason:   "When URL is empty, only name should be checked",
		},
		{
			name: "explicit fetch mode overrides patterns",
			node: NodeDefinition{
				Name:  "backend-meta",
				URL:   "https://github.com/org/backend-meta.git",
				Fetch: FetchLazy,
			},
			expected: true, // Should be lazy because explicitly set
			reason:   "Explicit fetch mode should override pattern detection",
		},
		{
			name: "case insensitive URL matching",
			node: NodeDefinition{
				Name:  "fulfillment",
				URL:   "https://github.com/org/FULFILLMENT-MUNOREPO.git",
				Fetch: FetchAuto,
			},
			expected: false, // Should be eager
			reason:   "URL matching should be case-insensitive",
		},
		{
			name: "renamed team-musinsa-munorepo to team-musinsa",
			node: NodeDefinition{
				Name:  "team-musinsa",
				URL:   "https://github.com/taokim/team-musinsa-munorepo.git",
				Fetch: FetchAuto,
			},
			expected: false, // Should be eager because URL matches pattern
			reason:   "Real-world case: renamed node should still be eager based on URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.node.IsLazy()
			assert.Equal(t, tt.expected, result, tt.reason)
		})
	}
}