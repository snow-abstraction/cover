package tree

import (
	"math"
)

type NodeKind byte

const (
	root NodeKind = 0
	// In the "both" branch subproblem the two branching
	// constraints should be covered by the same variable
	bothBranch = 1
	// In the "diff" branch subproblem the two branching
	// constrains should be covered by different variables
	diffBranch = 2
)

// constraint branch-and-bound node
// The subproblem the node represents can be calculated by applying
// the branch type of it and its ancestors.
type node struct {
	kind       NodeKind
	parent     *node // nil if root node
	lowerBound float64
	// The following have no meaning for the root node
	branchConstraintOne uint32
	branchConstraintTwo uint32
}

func createRoot() *node {
	return &node{root, nil, math.MaxFloat64, math.MaxUint32, math.MaxUint32}
}

func CreateInitialNodes() []*node {
	return []*node{createRoot()}
}

// Branches the parent on the two constrains to create two new Nodes
func (parent *node) Branch(lowerBound float64, branchConstraintOne uint32,
	branchConstraintTwo uint32) (*node, *node) {

	if parent == nil {
		panic("Cannot branch nil node.")
	}

	return &node{bothBranch, parent, lowerBound, branchConstraintOne, branchConstraintTwo},
		&node{diffBranch, parent, lowerBound, branchConstraintOne, branchConstraintTwo}

}
