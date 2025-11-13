package tree

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/andreygrechin/gitree/internal/models"
)

// FormatOptions configures tree formatting behavior
type FormatOptions struct {
	// ShowRoot determines whether to show the root directory in output
	ShowRoot bool

	// RootLabel is the label to use for the root (e.g., ".")
	RootLabel string
}

// DefaultFormatOptions returns sensible defaults
func DefaultFormatOptions() *FormatOptions {
	return &FormatOptions{
		ShowRoot:  true,
		RootLabel: ".",
	}
}

// Build constructs a hierarchical tree structure from a flat list of repositories
func Build(rootPath string, repos []*models.Repository, opts *FormatOptions) *models.TreeNode {
	if opts == nil {
		opts = DefaultFormatOptions()
	}

	// Create root node
	root := &models.TreeNode{
		Repository: &models.Repository{
			Path: rootPath,
			Name: opts.RootLabel,
		},
		Depth:        0,
		IsLast:       false,
		Children:     make([]*models.TreeNode, 0),
		RelativePath: opts.RootLabel,
	}

	if repos == nil {
		return root
	}

	// Build tree by organizing repos into hierarchy
	for _, repo := range repos {
		relPath, err := filepath.Rel(rootPath, repo.Path)
		if err != nil {
			// Skip repos that aren't under root
			continue
		}

		// Create or find nodes for this repository
		insertIntoTree(root, repo, relPath, rootPath)
	}

	// Sort all children alphabetically and mark IsLast flags
	sortTree(root)

	return root
}

// insertIntoTree inserts a repository into the tree at the correct location
func insertIntoTree(root *models.TreeNode, repo *models.Repository, relPath string, rootPath string) {
	parts := strings.Split(filepath.ToSlash(relPath), "/")

	current := root
	currentPath := rootPath

	// Navigate/create intermediate directory nodes
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		currentPath = filepath.Join(currentPath, part)

		// Find or create child node for this directory component
		found := false
		for _, child := range current.Children {
			if child.Repository.Name == part {
				current = child
				found = true
				break
			}
		}

		if !found {
			// Create intermediate directory node (not a git repo itself)
			newNode := &models.TreeNode{
				Repository: &models.Repository{
					Path: currentPath,
					Name: part,
				},
				Children:     make([]*models.TreeNode, 0),
				RelativePath: strings.Join(parts[:i+1], "/"),
			}
			current.Children = append(current.Children, newNode)
			current = newNode
		}
	}

	// Add the actual repository node
	repoNode := &models.TreeNode{
		Repository:   repo,
		Children:     make([]*models.TreeNode, 0),
		RelativePath: relPath,
	}
	current.Children = append(current.Children, repoNode)
}

// sortTree recursively sorts all children alphabetically and sets depth/IsLast flags
func sortTree(node *models.TreeNode) {
	if node == nil {
		return
	}

	// Sort children using the TreeNode.SortChildren method
	node.SortChildren()

	// Recursively sort all children and set their depth
	for _, child := range node.Children {
		child.Depth = node.Depth + 1
		sortTree(child)
	}
}

// Format generates ASCII tree output from a tree structure
func Format(root *models.TreeNode, opts *FormatOptions) string {
	if opts == nil {
		opts = DefaultFormatOptions()
	}

	var builder strings.Builder

	if root == nil {
		return ""
	}

	// If there are no children, return message
	if len(root.Children) == 0 {
		if opts.ShowRoot {
			builder.WriteString(fmt.Sprintf("%s\n", opts.RootLabel))
		}
		return builder.String()
	}

	// Show root if requested
	if opts.ShowRoot {
		builder.WriteString(fmt.Sprintf("%s\n", opts.RootLabel))
	}

	// Format children
	for i, child := range root.Children {
		isLast := (i == len(root.Children)-1)
		formatNode(&builder, child, "", isLast)
	}

	return builder.String()
}

// formatNode recursively formats a tree node with appropriate connectors
func formatNode(builder *strings.Builder, node *models.TreeNode, prefix string, isLast bool) {
	if node == nil || node.Repository == nil {
		return
	}

	// Choose connector based on whether this is the last child
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	// Write the node line
	builder.WriteString(prefix)
	builder.WriteString(connector)
	builder.WriteString(node.Repository.Name)

	// Add Git status if available
	if node.Repository.GitStatus != nil {
		builder.WriteString(" ")
		builder.WriteString(node.Repository.GitStatus.Format())
	}

	// Add error indicator if present
	if node.Repository.Error != nil && node.Repository.GitStatus == nil {
		builder.WriteString(" error")
	}

	// Add timeout indicator if present
	if node.Repository.HasTimeout && node.Repository.GitStatus != nil && node.Repository.GitStatus.Error != "" {
		builder.WriteString(" timeout")
	}

	// Add bare indicator if this is a bare repository
	if node.Repository.IsBare && node.Repository.GitStatus != nil {
		// Check if "bare" is already in the status format
		statusStr := node.Repository.GitStatus.Format()
		if !strings.Contains(statusStr, "bare") {
			builder.WriteString(" bare")
		}
	}

	builder.WriteString("\n")

	// Format children with updated prefix
	childPrefix := prefix
	if isLast {
		childPrefix += "    " // Four spaces for last child
	} else {
		childPrefix += "│   " // Vertical bar and three spaces for non-last
	}

	for i, child := range node.Children {
		childIsLast := (i == len(node.Children)-1)
		formatNode(builder, child, childPrefix, childIsLast)
	}
}
