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
	"errors"
	"fmt"
)

// For printing the implicit tree struct of Nodes
type printNode struct {
	referenceNode   *Node
	bothBranchChild *printNode
	diffBranchChild *printNode
}

// For the start node and its ancestors, create corresponding PrintNodes if
// they are not already in printNodeByNode. And set the links for the PrintNodes
// from the parent to its children.
func add(printNodeByNode map[*Node]*printNode, start *Node) (*printNode, error) {
	curr := start // curr = current
	var prev *Node

	var prevPNode *printNode
	var currPNode *printNode

	// Isn't necessary to actually follow parents to the root node
	// if a node is already in printNodeByNode, but we do so to
	// check for errors.
	for curr != nil {
		var ok bool
		currPNode, ok = printNodeByNode[curr]
		if !ok {
			currPNode = &printNode{curr, nil, nil}
			printNodeByNode[curr] = currPNode
		}

		if prev != nil {
			switch prev.Kind {
			case Root:
				return nil, fmt.Errorf("node of kind root has a non-nil parent %+v", *curr)
			case BothBranch:
				if currPNode.bothBranchChild == nil {
					currPNode.bothBranchChild = prevPNode
				} else if currPNode.bothBranchChild != prevPNode {
					return nil, fmt.Errorf(
						"bothBranchChild set before to a different node for node %+v", *curr)
				}
			case DiffBranch:
				if currPNode.diffBranchChild == nil {
					currPNode.diffBranchChild = prevPNode
				} else if currPNode.diffBranchChild != prevPNode {
					return nil, fmt.Errorf(
						"diffBranchChild set before to a different node for node %+v", *curr)
				}
			default:
				return nil, fmt.Errorf("unknown kind for for node %+v", *curr)

			}
		}

		prev = curr
		prevPNode = currPNode
		curr = curr.Parent

	}

	return currPNode, nil
}

func printImpl(depth int, node *printNode) {
	if node == nil {
		return
	}

	for i := 0; i < depth; i++ {
		fmt.Printf(" ")
	}

	fmt.Printf("%+v\n", *node.referenceNode)
	printImpl(depth+2, node.diffBranchChild)
	printImpl(depth+2, node.diffBranchChild)

}

// For the nodes, find all ancestors and print the tree of nodes
// All the supplied nodes, must have the same root.
func PrintTree(nodes []*Node) error {
	if len(nodes) == 0 {
		return nil
	}

	var root *printNode
	m := make(map[*Node]*printNode)
	for _, node := range nodes {
		r, err := add(m, node)
		if err != nil {
			return nil
		}

		if root != nil && r != root {
			return errors.New("two different root nodes found")
		}
		root = r
	}

	printImpl(0, root)
	return nil
}
