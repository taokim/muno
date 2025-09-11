package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdditionalCoverage(t *testing.T) {
	t.Run("GetDefaultWorkspaceName", func(t *testing.T) {
		name := GetDefaultWorkspaceName()
		assert.Equal(t, "muno-workspace", name)
	})
	
	t.Run("GetDefaultReposDir", func(t *testing.T) {
		dir := GetDefaultReposDir()
		assert.Equal(t, "repos", dir)
	})
	
	t.Run("GetStateFileName", func(t *testing.T) {
		name := GetStateFileName()
		assert.Equal(t, ".muno-tree.json", name)
	})
	
	t.Run("GetTreeDisplay", func(t *testing.T) {
		display := GetTreeDisplay()
		assert.NotEmpty(t, display.Indent)
		assert.NotEmpty(t, display.Branch)
		assert.NotEmpty(t, display.LastBranch)
		assert.NotEmpty(t, display.Vertical)
		assert.NotEmpty(t, display.Space)
	})
	
	t.Run("GetIcons", func(t *testing.T) {
		icons := GetIcons()
		assert.NotEmpty(t, icons.Cloned)
		assert.NotEmpty(t, icons.Lazy)
		assert.NotEmpty(t, icons.Modified)
	})
}