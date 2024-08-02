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

// bb = branch-and-bound

package solvers

import (
	"slices"

	"github.com/snow-abstraction/cover/internal/tree"
)

// createSubproblem creates new instance with only the subsets that are allowed by the
// constraints from the node and its ancestors. If some element is
// impossible to cover then it returns nil.
func createSubproblem(ins instance, node *tree.Node) *instance {

	type constraint struct {
		i            uint32
		j            uint32
		isBothBranch bool // if not, is different branch
	}

	costs := make([]float64, 0, len(ins.costs))
	subsets := make([][]int, 0, len(ins.subsets))
	isElementCovered := make([]bool, ins.m)

	constraints := make([]constraint, 0)
	for nodeI := node; nodeI.Kind != tree.Root; nodeI = nodeI.Parent {
		isBothBranch := nodeI.Kind == tree.BothBranch
		c := constraint{i: nodeI.I, j: nodeI.J, isBothBranch: isBothBranch}
		constraints = append(constraints, c)
	}

	for i, subset := range ins.subsets {
		noConstraintsViolated := true
		for _, c := range constraints {
			// TODO: check these int casts or eliminate them
			hasI := slices.Contains(subset, int(c.i))
			hasJ := slices.Contains(subset, int(c.j))
			if c.isBothBranch {
				if hasI != hasJ {
					noConstraintsViolated = false
					break
				}
			} else {
				if hasI && hasJ {
					noConstraintsViolated = false
					break
				}
			}
		}
		if noConstraintsViolated {
			costs = append(costs, ins.costs[i])
			subsets = append(subsets, subset)
			for _, e := range subset {
				// TODO: check these int casts or eliminate them
				idx := int(e)
				isElementCovered[idx] = true
			}
		}
	}

	for _, covered := range isElementCovered {
		if !covered {
			return nil
		}
	}

	return &instance{m: ins.m, subsets: subsets, costs: costs}

}
