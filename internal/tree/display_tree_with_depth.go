package tree

import (
	"fmt"
	"strings"
)

// DisplayTreeWithDepth shows tree with depth limit
func (m *StatelessManager) DisplayTreeWithDepth(maxDepth int) string {
	if maxDepth <= 0 {
		// Only show workspace name
		return fmt.Sprintf("🌳 %s\n", m.config.Workspace.Name)
	}
	
	// For depth 1, only show top-level nodes (not nested ones)
	var output strings.Builder
	output.WriteString(fmt.Sprintf("🌳 %s\n", m.config.Workspace.Name))
	
	// Filter nodes to only show those at the root level
	// In the flat structure, nested nodes would have been added with a parent path
	// but since we're storing them flat, we need a different approach
	// For now, we'll assume nodes without "/" in their name are top-level
	topLevelCount := 0
	for _, node := range m.config.Nodes {
		// Skip nodes that appear to be nested (contain path separators)
		if !strings.Contains(node.Name, "/") {
			topLevelCount++
		}
	}
	
	currentIndex := 0
	for _, node := range m.config.Nodes {
		// Skip nested nodes for depth 1
		if strings.Contains(node.Name, "/") {
			continue
		}
		
		prefix := "├─"
		if currentIndex == topLevelCount-1 {
			prefix = "└─"
		}
		currentIndex++
		
		icon := "📦"
		fsPath := m.ComputeFilesystemPath("/" + node.Name)
		state := GetRepoState(fsPath)
		
		if state == RepoStateMissing {
			icon = "💤"
		} else if state == RepoStateModified {
			icon = "📝"
		}
		
		if node.Config != "" {
			icon = "📁"
		}
		
		status := ""
		if node.Lazy && state == RepoStateMissing {
			status = " (lazy)"
		}
		
		output.WriteString(fmt.Sprintf("%s %s %s%s\n", prefix, icon, node.Name, status))
	}
	
	return output.String()
}