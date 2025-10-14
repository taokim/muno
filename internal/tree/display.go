package tree

import (
	"fmt"
	"path"
	"strings"
)

// DisplayTree returns a string representation of the entire tree
func (m *Manager) DisplayTree() string {
	return m.displayNode("/", "", true, 0, -1)
}

// DisplayTreeWithDepth returns a string representation of the tree up to a certain depth
func (m *Manager) DisplayTreeWithDepth(maxDepth int) string {
	return m.displayNode("/", "", true, 0, maxDepth)
}

// displayNode recursively builds the tree display
func (m *Manager) displayNode(logicalPath, prefix string, isLast bool, depth int, maxDepth int) string {
	// Check depth limit
	if maxDepth >= 0 && depth > maxDepth {
		return ""
	}
	
	node := m.GetNode(logicalPath)
	if node == nil {
		return ""
	}
	
	var sb strings.Builder
	
	// Don't show prefix for root
	if logicalPath != "/" {
		// Display current node
		symbol := "├─"
		if isLast {
			symbol = "└─"
		}
		
		// Choose icon based on node type and state
		icon := m.getNodeIcon(node)
		
		// Add current indicator
		current := ""
		if logicalPath == m.currentPath {
			current = " 📍"
		}
		
		// Show node info
		nodeInfo := node.Name
		if node.Type == NodeTypeRepo || node.Type == NodeTypeRepository {
			status := ""
			if node.State == RepoStateMissing {
				status = " (lazy)"
			} else if node.State == RepoStateModified {
				status = " (modified)"
			}
			nodeInfo += status
		}
		
		sb.WriteString(fmt.Sprintf("%s%s %s %s%s\n", prefix, symbol, icon, nodeInfo, current))
	} else {
		// For root, just show the workspace name
		workspaceName := "Workspace"
		if m.config != nil && m.config.Workspace.Name != "" {
			workspaceName = m.config.Workspace.Name
		}
		sb.WriteString(fmt.Sprintf("🌳 %s%s\n", workspaceName, m.getCurrentIndicator("/")))
	}
	
	// Display children
	childPrefix := prefix
	if logicalPath != "/" {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	}
	
	for i, childName := range node.Children {
		childPath := path.Join(logicalPath, childName)
		isLastChild := i == len(node.Children)-1
		sb.WriteString(m.displayNode(childPath, childPrefix, isLastChild, depth+1, maxDepth))
	}
	
	return sb.String()
}

// getNodeIcon returns an appropriate icon for the node
func (m *Manager) getNodeIcon(node *TreeNode) string {
	if node.Type == NodeTypeRoot {
		return "🌳"
	}
	
	if node.Type == NodeTypeRepo || node.Type == NodeTypeRepository {
		switch node.State {
		case RepoStateMissing:
			return "💤" // Lazy/not cloned
		case RepoStateModified:
			return "📝" // Modified
		default:
			return "📦" // Cloned
		}
	}
	
	return "📁" // Default folder icon
}

// getCurrentIndicator returns a current position indicator if this is the current node
func (m *Manager) getCurrentIndicator(logicalPath string) string {
	if logicalPath == m.currentPath {
		return " 📍"
	}
	return ""
}

// DisplayStatus shows the current state of the tree
func (m *Manager) DisplayStatus() string {
	var sb strings.Builder
	
	sb.WriteString("=== Tree Status ===\n")
	sb.WriteString(fmt.Sprintf("Current Path: %s\n", m.currentPath))
	sb.WriteString(fmt.Sprintf("Total Nodes: %d\n", len(m.config.Nodes)))
	
	// Count repositories by state
	var totalRepos, clonedRepos, lazyRepos, modifiedRepos int
	for _, nodeDef := range m.config.Nodes {
		node := m.GetNode("/" + nodeDef.Name)
		if node != nil && (node.Type == NodeTypeRepo || node.Type == NodeTypeRepository) {
			totalRepos++
			switch node.State {
			case RepoStateCloned:
				clonedRepos++
			case RepoStateMissing:
				lazyRepos++
			case RepoStateModified:
				modifiedRepos++
			}
		}
	}
	
	sb.WriteString(fmt.Sprintf("Repositories: %d total (%d cloned, %d lazy, %d modified)\n",
		totalRepos, clonedRepos, lazyRepos, modifiedRepos))
	
	// Show filesystem path for current node
	fsPath := m.ComputeFilesystemPath(m.currentPath)
	sb.WriteString(fmt.Sprintf("Filesystem Path: %s\n", fsPath))
	
	return sb.String()
}

// DisplayPath shows the path from root to current node
func (m *Manager) DisplayPath() string {
	parts := strings.Split(strings.TrimPrefix(m.currentPath, "/"), "/")
	if m.currentPath == "/" {
		return "/"
	}
	
	var path strings.Builder
	path.WriteString("/")
	
	for i, part := range parts {
		if i > 0 {
			path.WriteString(" → ")
		}
		path.WriteString(part)
	}
	
	return path.String()
}

// DisplayChildren shows just the immediate children of the current node
func (m *Manager) DisplayChildren() string {
	node := m.GetNode(m.currentPath)
	if node == nil {
		return "Current node not found"
	}
	
	if len(node.Children) == 0 {
		return "No children"
	}
	
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Children of %s:\n", m.currentPath))
	
	for _, childName := range node.Children {
		childPath := path.Join(m.currentPath, childName)
		child := m.GetNode(childPath)
		if child != nil {
			icon := m.getNodeIcon(child)
			status := ""
			if child.Type == NodeTypeRepo || child.Type == NodeTypeRepository {
				if child.State == RepoStateMissing {
					status = " (lazy)"
				} else if child.State == RepoStateModified {
					status = " (modified)"
				}
			}
			sb.WriteString(fmt.Sprintf("  %s %s%s\n", icon, childName, status))
		}
	}
	
	return sb.String()
}