/*
 Copyright (C) 2024 Douglas Wayne Potter

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU Affero General Public License as
 published by the Free Software Foundation, either version 3 of the
 License, or (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Affero General Public License for more details.

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package tree

import (
	"fmt"
	"log/slog"
	"math"
)

type NodeKind byte

const (
	Root NodeKind = 0
	// In the "both" branch subproblem the two branching
	// constraints should be covered by the same variable
	BothBranch = 1
	// In the "diff" branch subproblem the two branching
	// constraints should be covered by different variables
	DiffBranch = 2
)

// constraint branch-and-bound Node
// The subproblem the Node represents can be calculated by applying
// the branch type of it and its ancestors.
// TODO: enforce I < J
type Node struct {
	Kind       NodeKind
	Parent     *Node // nil if root node
	LowerBound float64
	// Elements constrained, depending on NodeKind
	// The following have no meaning for the root node
	I uint32
	J uint32
}

func CreateRoot() *Node {
	return &Node{Root, nil, math.MaxFloat64, math.MaxUint32, math.MaxUint32}
}

func CreateInitialNodes() []*Node {
	return []*Node{CreateRoot()}
}

// Branches the parent on the two constrains to create two new Nodes
func (parent *Node) Branch(lowerBound float64, branchConstraintOne uint32,
	branchConstraintTwo uint32) (*Node, *Node) {

	return &Node{BothBranch, parent, lowerBound, branchConstraintOne, branchConstraintTwo},
		&Node{DiffBranch, parent, lowerBound, branchConstraintOne, branchConstraintTwo}

}

func (n *Node) LogValue() slog.Value {
	if n == nil {
		return slog.StringValue("nil")
	}

	return slog.StringValue(fmt.Sprintf("%+v", *n))
}
