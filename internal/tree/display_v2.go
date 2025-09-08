package tree

import (
	"fmt"
	"io"
	"strings"

	"github.com/taokim/muno/internal/tree/navigator"
)

// DisplayOptions configures display output
type DisplayOptions struct {
	ShowStatus bool
	ShowLazy   bool
	ShowBranch bool
	ShowURL    bool
	Verbose    bool
}

// Display handles tree and status visualization
type Display struct {
	writer io.Writer
	opts   *DisplayOptions
}

// NewDisplay creates a new display handler
func NewDisplay(w io.Writer, opts *DisplayOptions) *Display {
	if opts == nil {
		opts = &DisplayOptions{}
	}
	return &Display{
		writer: w,
		opts:   opts,
	}
}

// PrintTreeView prints a tree view to the writer
func (d *Display) PrintTreeView(tree *navigator.TreeView) error {
	if tree == nil || tree.Root == nil {
		return fmt.Errorf("invalid tree view")
	}

	fmt.Fprintf(d.writer, "Tree from %s:\n", tree.Root.Path)
	return d.printNode(tree, tree.Root, "", true)
}

// PrintStatusView prints a status view to the writer
func (d *Display) PrintStatusView(tree *navigator.TreeView) error {
	if tree == nil {
		return fmt.Errorf("invalid tree view")
	}

	fmt.Fprintln(d.writer, "Repository Status:")
	fmt.Fprintln(d.writer, strings.Repeat("-", 60))

	for path, node := range tree.Nodes {
		if node.Type == navigator.NodeTypeRepo {
			status := tree.Status[path]
			if status == nil {
				continue
			}

			// Build status line
			var statusStr string
			switch status.State {
			case navigator.RepoStateMissing:
				statusStr = "[ MISSING ]"
			case navigator.RepoStateCloned:
				if status.Modified {
					statusStr = "[MODIFIED ]"
				} else {
					statusStr = "[ CLEAN   ]"
				}
			case navigator.RepoStateModified:
				statusStr = "[MODIFIED ]"
			case navigator.RepoStateAhead:
				statusStr = "[ AHEAD   ]"
			case navigator.RepoStateBehind:
				statusStr = "[ BEHIND  ]"
			case navigator.RepoStateDiverged:
				statusStr = "[DIVERGED ]"
			default:
				statusStr = "[ UNKNOWN ]"
			}

			// Add branch info if requested
			branchInfo := ""
			if d.opts.ShowBranch && status.Branch != "" {
				branchInfo = fmt.Sprintf(" (%s)", status.Branch)
			}

			// Add lazy indicator
			lazyInfo := ""
			if d.opts.ShowLazy && status.Lazy {
				lazyInfo = " [lazy]"
			}

			fmt.Fprintf(d.writer, "%s %s%s%s\n", statusStr, path, branchInfo, lazyInfo)

			// Show URL if requested
			if d.opts.ShowURL && node.URL != "" {
				fmt.Fprintf(d.writer, "         URL: %s\n", node.URL)
			}

			// Show error if any
			if status.Error != "" {
				fmt.Fprintf(d.writer, "         Error: %s\n", status.Error)
			}
		}
	}

	return nil
}

// PrintNode prints a single node with details
func (d *Display) PrintNode(node *navigator.Node, status *navigator.NodeStatus) error {
	if node == nil {
		return fmt.Errorf("node is nil")
	}

	fmt.Fprintf(d.writer, "Node: %s\n", node.Path)
	fmt.Fprintf(d.writer, "  Name: %s\n", node.Name)
	fmt.Fprintf(d.writer, "  Type: %s\n", node.Type)

	if node.URL != "" {
		fmt.Fprintf(d.writer, "  URL: %s\n", node.URL)
	}

	if node.ConfigRef != "" {
		fmt.Fprintf(d.writer, "  Config: %s\n", node.ConfigRef)
	}

	if len(node.Children) > 0 {
		fmt.Fprintf(d.writer, "  Children: %s\n", strings.Join(node.Children, ", "))
	}

	if status != nil && d.opts.ShowStatus {
		fmt.Fprintln(d.writer, "  Status:")
		fmt.Fprintf(d.writer, "    Exists: %v\n", status.Exists)
		fmt.Fprintf(d.writer, "    Cloned: %v\n", status.Cloned)
		fmt.Fprintf(d.writer, "    State: %s\n", status.State)
		if status.Modified {
			fmt.Fprintf(d.writer, "    Modified: true\n")
		}
		if status.Lazy {
			fmt.Fprintf(d.writer, "    Lazy: true\n")
		}
		if status.Branch != "" {
			fmt.Fprintf(d.writer, "    Branch: %s\n", status.Branch)
		}
		if status.Error != "" {
			fmt.Fprintf(d.writer, "    Error: %s\n", status.Error)
		}
	}

	return nil
}

// Helper method for tree printing
func (d *Display) printNode(tree *navigator.TreeView, node *navigator.Node, prefix string, isLast bool) error {
	// Determine the connector
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	// Build node display
	nodeStr := node.Name
	
	// Add type indicator
	switch node.Type {
	case navigator.NodeTypeRepo:
		if status := tree.Status[node.Path]; status != nil {
			if status.Lazy && !status.Cloned {
				nodeStr += " [lazy]"
			} else if status.Modified {
				nodeStr += " [M]"
			}
		}
		nodeStr += " (repo)"
	case navigator.NodeTypeConfig:
		nodeStr += " (config)"
	case navigator.NodeTypeDirectory:
		nodeStr += " (dir)"
	}

	// Print the node
	fmt.Fprintf(d.writer, "%s%s%s\n", prefix, connector, nodeStr)

	// Update prefix for children
	childPrefix := prefix
	if isLast {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}

	// Print children
	for i, childName := range node.Children {
		childPath := node.Path + "/" + childName
		if node.Path == "/" {
			childPath = "/" + childName
		}

		if childNode, exists := tree.Nodes[childPath]; exists {
			isLastChild := (i == len(node.Children)-1)
			if err := d.printNode(tree, childNode, childPrefix, isLastChild); err != nil {
				return err
			}
		}
	}

	return nil
}

// PrintPath prints the current path
func (d *Display) PrintPath(path string) error {
	fmt.Fprintf(d.writer, "Current path: %s\n", path)
	return nil
}

// PrintChildren prints a list of children
func (d *Display) PrintChildren(children []*navigator.Node) error {
	if len(children) == 0 {
		fmt.Fprintln(d.writer, "No children")
		return nil
	}

	fmt.Fprintf(d.writer, "Children (%d):\n", len(children))
	for _, child := range children {
		typeStr := string(child.Type)
		if child.Type == navigator.NodeTypeRepo && child.URL != "" {
			typeStr = "repo"
		}
		fmt.Fprintf(d.writer, "  - %s (%s)\n", child.Name, typeStr)
	}

	return nil
}