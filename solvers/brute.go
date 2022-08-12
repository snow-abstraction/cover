/*
 Copyright (C) 2022 Douglas Wayne Potter

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

// A brute force solver for the "Set Partitioning Problem".
package solvers

import (
	"gonum.org/v1/gonum/stat/combin"
)

// Attempt to make a solution by adding the candidates (subsets) one by one (in order)
// until either feasible or infeasible (overcovered or no subsets left).
//
// Returns the solution (the subset indices) and its cost. If no solution, then
// it returns nil and NaN
func makeSolutionFromSubsets(ins instance, subsetIndices []int) subsetsEval {
	coverCounts := make([]int, ins.n)

	var s subsetsEval

	if ins.n == 0 {
		s.exactlyCovered = true
		return s
	}

	for _, subsetIdx := range subsetIndices {
		for _, elementIdx := range ins.subsets[subsetIdx] {
			coverCounts[elementIdx] += 1
		}

		s.cost += ins.costs[subsetIdx]
		// TODO: the use case is find a solution (aka a partition) that is cheaper than
		// a known solution so we could abort the search if s.cost greater than the cost
		// of a known partition.
		s.subsetsIndices = append(s.subsetsIndices, subsetIdx)

		allConstraintsCoveredExactly := true
		for _, coverCount := range coverCounts {
			isOverCovered := coverCount > 1
			if isOverCovered {
				return s
			} else if coverCount == 0 {
				allConstraintsCoveredExactly = false
			}
		}

		if allConstraintsCoveredExactly {
			s.exactlyCovered = true
			return s
		}
	}

	return s
}

func SolveByBruteForce(ins instance) subsetsEval {

	nSubsetsToTry := ins.n
	if len(ins.subsets) < ins.n {
		nSubsetsToTry = len(ins.subsets)
	}

	permutations := combin.NewPermutationGenerator(len(ins.subsets), nSubsetsToTry)
	if permutations.Next() == false {
		panic("No permutations.")
	}
	bestSubsetsEval := makeSolutionFromSubsets(ins, permutations.Permutation(nil))

	for permutations.Next() {
		perm := permutations.Permutation(nil)
		subsetEval := makeSolutionFromSubsets(ins, perm)
		if !subsetEval.exactlyCovered {
			continue
		} else if !bestSubsetsEval.exactlyCovered || subsetEval.cost < bestSubsetsEval.cost {
			bestSubsetsEval = subsetEval
		}

	}

	return bestSubsetsEval
}
