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
	"fmt"
	"log/slog"
	"slices"

	"github.com/snow-abstraction/cover/internal/tree"
)

type subInstance struct {
	ins     instance
	indices []int // indices[i] == index in original problem
	// If ins is a solution itself, that is the subsets of instance are an exact cover
	isSolution bool
}

// createSubInstance creates new instance with only the subsets that are allowed by the
// constraints from the node and its ancestors. If some element is
// impossible to cover then it returns nil.
func createSubInstance(ins instance, node *tree.Node) *subInstance {

	type constraint struct {
		i            uint32
		j            uint32
		isBothBranch bool // if not, is different branch
	}

	costs := make([]float64, 0, len(ins.costs))
	subsets := make([][]int, 0, len(ins.subsets))
	// count of the subsets in sub-instance
	coverCount := make([]int, ins.m)
	indices := make([]int, 0, len(ins.subsets))

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
			indices = append(indices, i)
			costs = append(costs, ins.costs[i])
			subsets = append(subsets, subset)
			for _, e := range subset {
				// TODO: check these int casts or eliminate them
				idx := int(e)
				coverCount[idx]++
			}
		}
	}

	isSolution := true // If the sub-instance is a solution.
	for _, covered := range coverCount {
		if covered == 0 {
			return nil
		} else if covered > 1 {
			isSolution = false
		}
	}

	return &subInstance{
		ins:        instance{m: ins.m, subsets: subsets, costs: costs},
		indices:    indices,
		isSolution: isSolution,
	}

}

type solution struct {
	objectiveValue float64
	subsetIndices  []int
}

func sum(xs []float64) float64 {
	total := 0.0
	for _, x := range xs {
		total += x
	}
	return total
}

// WIP
func SolveByBranchAndBound(ins instance) (subsetsEval, error) {
	if ins.m == 0 {
		return subsetsEval{
			ExactlyCovered: true,
			Optimal:        true,
		}, nil
	}

	var best *solution
	toFathom := make([]*tree.Node, 0)
	root := tree.CreateRoot()
	toFathom = append(toFathom, root)

	for len(toFathom) > 0 {
		node := toFathom[0]
		toFathom = toFathom[1:]
		slog.Debug("B&B status", "nodes count", len(toFathom), "node", node)

		if best != nil && best.objectiveValue <= node.LowerBound {
			slog.Debug("discarding node", "node", node, "best obj val", best.objectiveValue)
			// discard node due to lower bound
			continue
		}

		subInstance := createSubInstance(ins, node)
		if subInstance == nil {
			slog.Debug("sub-instance infeasible")
			continue
		} else if subInstance.isSolution {
			cost := sum(subInstance.ins.costs)
			if best == nil || best.objectiveValue > cost {
				best = &solution{cost, subInstance.indices}
				slog.Debug("new best solution from sub-instance", "solution", best)
			}
			continue
		}

		// Here we know that subInstance either has no solution or has
		// non-trivial solution in the sense at least element is in two
		// or more subsets.
		matrix, err := convertSubsetsToMatrix(subInstance.ins.subsets)
		if err != nil {
			return subsetsEval{}, err
		}

		dualResult, err := runDualIterations(matrix, subInstance.ins.costs)
		if err != nil {
			return subsetsEval{}, err
		}
		if dualResult.provenOptimalExact {
			slog.Debug("pruned by optimal")
			if best == nil || best.objectiveValue > dualResult.dualObjectiveValue {
				indices := make([]int, 0, len(dualResult.primalSolution))
				for _, idx := range dualResult.primalSolution {
					indices = append(indices, subInstance.indices[idx])
				}
				best = &solution{dualResult.dualObjectiveValue, indices}
				slog.Debug("new best solution", "solution", best)
			}
			continue
		}

		if best != nil && best.objectiveValue <= dualResult.dualObjectiveValue {
			// We could only do this when getting the node.
			slog.Debug("pruned by bound", "node", node)
			continue
		}

		branchIndices, err := findBranchingElements(subInstance.ins)
		if err != nil {
			return subsetsEval{}, err
		}

		slog.Debug("branching on elements", "i", branchIndices.i, "j", branchIndices.j)
		bothNode, diffNode := node.Branch(dualResult.dualObjectiveValue, branchIndices.i, branchIndices.j)
		toFathom = append(toFathom, bothNode, diffNode)
	}

	// TODO: return non-optimal solutions
	if best == nil || len(toFathom) != 0 {
		return subsetsEval{}, nil
	}

	return subsetsEval{
		SubsetsIndices: best.subsetIndices,
		ExactlyCovered: true,
		Cost:           best.objectiveValue,
		Optimal:        true,
	}, nil
}

type BranchIndices struct {
	// elements to branch on
	i uint32
	j uint32
}

// findBranchingElements finds a pair of elements (i, j) that are covered both
// in same subset and different subsets. For instance, if we have subsets
// [0, 1], [1, 2], [2] then i=1 and j=2 would work.
//
// Assume it is possible to branch on the instance, i.e. should not be empty and
// it is not a solution itself.
func findBranchingElements(ins instance) (BranchIndices, error) {
	counts := make([]int, ins.m)
	for _, subset := range ins.subsets {
		for _, el := range subset {
			counts[el]++
		}
	}

	// i is the index of the two elements to branch on and is index of the element in the most subsets
	i := 0
	for idx, c := range counts {
		if counts[i] < c {
			i = idx
		}
	}

	if counts[i] <= 1 {
		return BranchIndices{},
			fmt.Errorf("at least one element must be in two subsets to branch on instance %+v", ins)
	}

	// For each element, count how many subsets does it share with maxIdx
	sharedCounts := make([]int, ins.m)
	for _, subset := range ins.subsets {
		// assume so few elements that faster with linear search
		if slices.Contains(subset, i) {
			for _, el := range subset {
				sharedCounts[el]++
			}
		}
	}

	// assumption checking: it should be possible to branch if guessBranchingElements is called.
	branchingPossible := false
	for j, count := range sharedCounts {
		if j == i {
			continue
		}
		if 0 < count && count < sharedCounts[i] {
			branchingPossible = true
		}
	}

	if !branchingPossible {
		return BranchIndices{},
			fmt.Errorf(
				"failed to find other element to branch on with %d for instance %+v",
				i, ins,
			)
	}

	var minValue int
	target := sharedCounts[i] / 2
	sharedCounts[i] = -1 // prevent choosing minIndex == maxIdx

	// The idea is select j such that is closer to target for more balanced
	// subproblems.
	j := -1
	for idx, count := range sharedCounts {
		distFromTarget := count - target
		distFromTarget = max(distFromTarget, -distFromTarget) //abs
		if j == -1 || distFromTarget < minValue {
			j = idx
			minValue = distFromTarget
		}
	}

	// TODO: check casts or remove need
	return BranchIndices{uint32(i), uint32(j)}, nil
}

// TODO: this method has some good ideas for branch. Revisit it.
// func findOtherElementToBranchOn(subproblemInstance instance, i int) uint32 {
// 	// For each element, count how many subsets does it share with i
// 	coverCounts := make([]int, subproblemInstance.m)
// 	for _, subset := range subproblemInstance.subsets {
// 		// assume so few elements that faster with linear search
// 		if slices.Contains(subset, i) {
// 			for _, el := range subset {
// 				coverCounts[el]++
// 			}
// 		}
// 	}

// 	// assumption checking: it should be possible to branch if findOtherElementToBranchOn is called.
// 	branchingPossible := false
// 	for j, count := range coverCounts {
// 		if j == i {
// 			continue
// 		}
// 		if 0 < count && count < coverCounts[i] {
// 			branchingPossible = true
// 		}
// 	}

// 	// TODO: return error
// 	if !branchingPossible {
// 		log.Fatalf(
// 			"failed to find other element to branch on with %d",
// 			i,
// 		)
// 	}

// 	var minValue int
// 	minIndex := -1
// 	// closer to target means more balanced subproblems
// 	target := coverCounts[i] / 2
// 	for j, count := range coverCounts {
// 		distFromTarget := count - target
// 		distFromTarget = max(distFromTarget, -distFromTarget) //abs
// 		if minIndex == -1 || distFromTarget < minValue {
// 			minIndex = j
// 			minValue = distFromTarget
// 		}
// 	}

// 	return uint32(minIndex)
// }
