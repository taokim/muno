package tree

import (
	"fmt"
	"strings"
)

// DisplayTree displays the tree structure
func (m *Manager) DisplayTree(node *Node, options TreeDisplay) string {
	if node == nil {
		node = m.rootNode
	}
	
	var sb strings.Builder
	m.displayNode(&sb, node, "", true, 0, options)
	return sb.String()
}

// DisplayList displays child nodes as a list
func (m *Manager) DisplayList(node *Node, recursive bool) string {
	if node == nil {
		node = m.currentNode
		if node == nil {
			node = m.rootNode
		}
	}
	
	var sb strings.Builder
	
	// Show current position
	sb.WriteString(fmt.Sprintf("📍 Current: %s\n", node.Path))
	
	// List children
	if len(node.Children) > 0 {
		sb.WriteString("\n📁 Child Nodes:\n")
		for name, child := range node.Children {
			repoCount := child.CountRepos(false)
			lazyCount := 0
			for _, repo := range child.Repos {
				if repo.Lazy && repo.State == string(RepoStateMissing) {
					lazyCount++
				}
			}
			
			status := "✓"
			if lazyCount > 0 {
				status = fmt.Sprintf("💤 %d lazy", lazyCount)
			}
			
			sb.WriteString(fmt.Sprintf("  • %s (%d repos, %s)\n", name, repoCount, status))
			
			if recursive && len(child.Children) > 0 {
				m.listChildrenRecursive(&sb, child, "    ")
			}
		}
	} else {
		sb.WriteString("\n  (no child nodes)\n")
	}
	
	// List repositories at this node
	if len(node.Repos) > 0 {
		sb.WriteString("\n📦 Repositories:\n")
		for _, repo := range node.Repos {
			status := "✓ cloned"
			if repo.State == string(RepoStateMissing) {
				if repo.Lazy {
					status = "💤 lazy"
				} else {
					status = "❌ missing"
				}
			} else if repo.State == string(RepoStateModified) {
				status = "📝 modified"
			}
			
			sb.WriteString(fmt.Sprintf("  • %s [%s]\n", repo.Name, status))
		}
	}
	
	return sb.String()
}

// DisplayStatus shows the status of a node and its repos
func (m *Manager) DisplayStatus(node *Node, recursive bool) string {
	if node == nil {
		node = m.currentNode
		if node == nil {
			node = m.rootNode
		}
	}
	
	var sb strings.Builder
	
	// Header
	sb.WriteString(fmt.Sprintf("📊 Status: %s\n", node.Path))
	sb.WriteString(strings.Repeat("─", 50) + "\n")
	
	// Node info
	sb.WriteString(fmt.Sprintf("Type: %s\n", node.Meta.Type))
	sb.WriteString(fmt.Sprintf("Repos: %d direct, %d total\n", 
		len(node.Repos), node.CountRepos(true)))
	
	// Repository status
	if len(node.Repos) > 0 {
		sb.WriteString("\n📦 Repository Status:\n")
		for _, repo := range node.Repos {
			m.displayRepoStatus(&sb, &repo, "  ")
		}
	}
	
	// Recursive status
	if recursive && len(node.Children) > 0 {
		sb.WriteString("\n📁 Child Nodes:\n")
		for name, child := range node.Children {
			stats := m.getNodeStats(child)
			sb.WriteString(fmt.Sprintf("  • %s: %d repos (%d cloned, %d lazy, %d modified)\n",
				name, stats.total, stats.cloned, stats.lazy, stats.modified))
		}
	}
	
	return sb.String()
}

// Helper functions

func (m *Manager) displayNode(sb *strings.Builder, node *Node, prefix string, isLast bool, depth int, options TreeDisplay) {
	if options.MaxDepth > 0 && depth > options.MaxDepth {
		return
	}
	
	// Node symbol
	nodeSymbol := "├─"
	if isLast {
		nodeSymbol = "└─"
	}
	if node.IsRoot() {
		nodeSymbol = ""
	}
	
	// Node type indicator
	typeIndicator := "📁"
	if node.IsLeaf() && len(node.Repos) > 0 {
		typeIndicator = "📦"
	}
	
	// Node name with info
	nodeInfo := node.Name
	if node.IsRoot() {
		nodeInfo = fmt.Sprintf("● %s (root)", node.Name)
	} else if options.ShowRepos {
		repoCount := len(node.Repos)
		if repoCount > 0 {
			nodeInfo = fmt.Sprintf("%s %s (%d repos)", typeIndicator, node.Name, repoCount)
		} else {
			nodeInfo = fmt.Sprintf("%s %s", typeIndicator, node.Name)
		}
	} else {
		nodeInfo = fmt.Sprintf("%s %s", typeIndicator, node.Name)
	}
	
	// Write node line
	if !node.IsRoot() {
		sb.WriteString(prefix)
	}
	sb.WriteString(nodeSymbol)
	if !node.IsRoot() {
		sb.WriteString(" ")
	}
	sb.WriteString(nodeInfo)
	
	// Show lazy indicator
	if options.ShowLazy && node.HasLazyRepos() {
		sb.WriteString(" 💤")
	}
	
	sb.WriteString("\n")
	
	// Show repositories if requested
	if options.ShowRepos && len(node.Repos) > 0 {
		repoPrefix := prefix
		if isLast {
			repoPrefix += "    "
		} else {
			repoPrefix += "│   "
		}
		
		for i, repo := range node.Repos {
			isLastRepo := i == len(node.Repos)-1 && len(node.Children) == 0
			m.displayRepo(sb, &repo, repoPrefix, isLastRepo, options)
		}
	}
	
	// Display children
	childPrefix := prefix
	if !node.IsRoot() {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	}
	
	i := 0
	for _, child := range node.Children {
		i++
		isLastChild := i == len(node.Children)
		m.displayNode(sb, child, childPrefix, isLastChild, depth+1, options)
	}
}

func (m *Manager) displayRepo(sb *strings.Builder, repo *RepoConfig, prefix string, isLast bool, options TreeDisplay) {
	symbol := "├─"
	if isLast {
		symbol = "└─"
	}
	
	status := ""
	if repo.State == string(RepoStateMissing) {
		if repo.Lazy {
			status = " [lazy]"
		} else {
			status = " [missing]"
		}
	} else if repo.State == string(RepoStateModified) && options.ShowModified {
		status = " [modified]"
	}
	
	sb.WriteString(fmt.Sprintf("%s%s 📄 %s%s\n", prefix, symbol, repo.Name, status))
}

func (m *Manager) displayRepoStatus(sb *strings.Builder, repo *RepoConfig, indent string) {
	status := "✓ cloned"
	if repo.State == string(RepoStateMissing) {
		if repo.Lazy {
			status = "💤 lazy (not cloned)"
		} else {
			status = "❌ missing"
		}
	} else if repo.State == string(RepoStateModified) {
		status = "📝 modified"
	}
	
	sb.WriteString(fmt.Sprintf("%s%s: %s\n", indent, repo.Name, status))
	
	if repo.State == string(RepoStateCloned) || repo.State == string(RepoStateModified) {
		// Could add git status info here
		sb.WriteString(fmt.Sprintf("%s  Path: %s\n", indent, repo.Path))
	}
}

func (m *Manager) listChildrenRecursive(sb *strings.Builder, node *Node, indent string) {
	for name, child := range node.Children {
		repoCount := child.CountRepos(false)
		sb.WriteString(fmt.Sprintf("%s• %s (%d repos)\n", indent, name, repoCount))
		
		if len(child.Children) > 0 {
			m.listChildrenRecursive(sb, child, indent+"  ")
		}
	}
}

type nodeStats struct {
	total    int
	cloned   int
	lazy     int
	modified int
}

func (m *Manager) getNodeStats(node *Node) nodeStats {
	stats := nodeStats{}
	
	for _, repo := range node.Repos {
		stats.total++
		switch repo.State {
		case string(RepoStateCloned):
			stats.cloned++
		case string(RepoStateModified):
			stats.modified++
		case string(RepoStateMissing):
			if repo.Lazy {
				stats.lazy++
			}
		}
	}
	
	// Include children stats
	for _, child := range node.Children {
		childStats := m.getNodeStats(child)
		stats.total += childStats.total
		stats.cloned += childStats.cloned
		stats.lazy += childStats.lazy
		stats.modified += childStats.modified
	}
	
	return stats
}