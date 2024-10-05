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
	"cmp"
	"fmt"
	"log/slog"
	"slices"

	"github.com/snow-abstraction/cover/internal/solvers/queue"
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
func createSubInstance(ins instance, node *tree.Node) (*subInstance, error) {
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

	branchedIndices := make(map[BranchIndices]struct{})

	constraints := make([]constraint, 0)
	for nodeI := node; nodeI.Kind != tree.Root; nodeI = nodeI.Parent {
		isBothBranch := nodeI.Kind == tree.BothBranch
		c := constraint{i: nodeI.I, j: nodeI.J, isBothBranch: isBothBranch}
		b := BranchIndices{nodeI.I, nodeI.J}
		if b.i >= b.j {
			return nil, fmt.Errorf("should have i < j but %+v", b)
		}
		if _, found := branchedIndices[b]; !found {
			branchedIndices[b] = struct{}{}
		} else {
			return nil, fmt.Errorf("already branched on %+v", b)
		}
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
			return nil, nil
		} else if covered > 1 {
			isSolution = false
		}
	}

	return &subInstance{
		ins:        instance{m: ins.m, subsets: subsets, costs: costs},
		indices:    indices,
		isSolution: isSolution,
	}, nil

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

// removeMoreExpensiveDuplicates returns a copy of instance ins
// with more expensive duplicates removed and originalIndexMap []int
// that maps indices of the new instance back ins. That is:
// ins.subset[originalIndexMap[i[]] == insCopy.subset[i].
func removeMoreExpensiveDuplicates(ins instance) (instance, []int) {
	type SubsetCostIndex struct {
		subset []int
		cost   float64
		index  int
	}

	temp := make([]SubsetCostIndex, 0, len(ins.subsets))
	for i := 0; i < len(ins.subsets); i++ {
		temp = append(temp, SubsetCostIndex{ins.subsets[i], ins.costs[i], i})
	}

	slices.SortFunc(temp, func(lhs, rhs SubsetCostIndex) int {
		if c := slices.Compare(lhs.subset, rhs.subset); c != 0 {
			return c
		}
		// Make lowest cost duplicate subset is first.
		return cmp.Compare(lhs.cost, rhs.cost)
	})

	ins.subsets = make([][]int, 0, len(ins.subsets))
	ins.costs = make([]float64, 0, len(ins.costs))
	originalIndexMap := make([]int, 0, len(ins.subsets))

	for i := 0; i < len(temp); {
		x := temp[i]
		ins.subsets = append(ins.subsets, x.subset)
		ins.costs = append(ins.costs, x.cost)
		originalIndexMap = append(originalIndexMap, x.index)

		i++
		for i < len(temp) && slices.Equal(x.subset, temp[i].subset) {
			i++
		}
	}

	return ins, originalIndexMap
}

// WIP
func SolveByBranchAndBound(ins instance) (subsetsEval, error) {
	if ins.m == 0 {
		return subsetsEval{
			ExactlyCovered: true,
			Optimal:        true,
		}, nil
	}

	// More expensive duplicates should never been in an optima and branching
	// scheme does not support duplicates.
	ins, originalIndexMap := removeMoreExpensiveDuplicates(ins)

	var best *solution
	toFathom := queue.MakeQueue()
	root := tree.CreateRoot()
	toFathom.Push(root)

	for toFathom.Len() > 0 {
		node := toFathom.Pop()
		slog.Debug("B&B status", "nodes count", toFathom.Len(), "node", node)

		if best != nil && best.objectiveValue <= node.LowerBound {
			slog.Debug("discarding node", "node", node, "best obj val", best.objectiveValue)
			// discard node due to lower bound
			continue
		}

		subInstance, err := createSubInstance(ins, node)
		if err != nil {
			return subsetsEval{}, err
		}
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
				best = &solution{dualResult.dualObjectiveValue,
					mapIndices(dualResult.primalSolution, subInstance.indices)}
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
		toFathom.Push(bothNode)
		toFathom.Push(diffNode)

	}

	// TODO: return non-optimal solutions
	if best == nil || toFathom.Len() != 0 {
		return subsetsEval{}, nil
	}

	// map indices back original instance indices
	indices := mapIndices(best.subsetIndices, originalIndexMap)

	return subsetsEval{
		SubsetsIndices: indices,
		ExactlyCovered: true,
		Cost:           best.objectiveValue,
		Optimal:        true,
	}, nil
}

func mapIndices(indices []int, indexMap []int) []int {
	mappedIndices := make([]int, 0, len(indices))
	for _, idx := range indices {
		mappedIndices = append(mappedIndices, indexMap[idx])
	}
	return mappedIndices
}

type BranchIndices struct {
	// elements to branch on
	i uint32
	j uint32
}

// symmetricDifference calculates the set symmetric difference of x and y.
func symmetricDifference(x, y []int) []int {
	xSet := make(map[int]struct{}, len(x))
	for _, e := range x {
		xSet[e] = struct{}{}
	}
	ySet := make(map[int]struct{}, len(y))
	for _, e := range y {
		ySet[e] = struct{}{}
	}

	diffSet := make(map[int]struct{}, len(x)+len(y))
	for _, e := range x {
		if _, found := ySet[e]; !found {
			diffSet[e] = struct{}{}
		}
	}
	for _, e := range y {
		if _, found := xSet[e]; !found {
			diffSet[e] = struct{}{}
		}
	}

	diff := make([]int, 0, len(diffSet))
	for k := range diffSet {
		diff = append(diff, k)
	}
	return diff
}

// findBranchingElements finds a pair of elements (i, j) that are covered both
// in same subset and different subsets. For instance, if we have subsets
// [0, 1], [1, 2], [2] then i=1 and j=2 would work.
//
// Assume it is possible to branch on the instance, i.e. should not be empty and
// it is not a solution itself. This means at least one element is in more than
// one subset.
//
// The implementation is unsophisticated but hopefully correct.
// It finds an element i that in the most subsets. Then it finds first two subsets
// containing i. Then it determine an element j such that is in only one
// of these two subsets. This means that these two subsets will be if different
// branches when branching on i and j.
//
// This is unsophisticated because:
//  1. It does not utilize data from running subgradient algorithm `runDualIterations`.
//     For example, by creating a branches that would not allowed the solution found
//     by the subgradient algorithm.
//  2. It does not try to find (i, j) that would result in smaller sub-instances.
//  3. When several choices are possible, the first is taken making it sensitive
//     to the ordering of the input.
func findBranchingElements(ins instance) (BranchIndices, error) {
	counts := make([]int, ins.m)
	for _, subset := range ins.subsets {
		for _, el := range subset {
			counts[el]++
		}
	}

	// i is the first index of the two elements to branch on and is index of the element in the most subsets
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

	// find two subsets with element i
	var subset1WithI []int
	var subset2WithI []int
	for _, subset := range ins.subsets {
		// subset is sorted so we could use binary search
		if !slices.Contains(subset, i) {
			continue
		}
		if subset1WithI == nil {
			subset1WithI = subset
		} else {
			subset2WithI = subset
			break
		}
	}

	diff := symmetricDifference(subset1WithI, subset2WithI)
	// The diff should not be empty since the subsets are unique.
	if len(diff) == 0 {
		return BranchIndices{},
			fmt.Errorf("branching failed for instance %+v", ins)
	}

	j := diff[0]

	if j < i {
		i, j = j, i
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
