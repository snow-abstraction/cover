/*
 Copyright (C) 2023 Douglas Wayne Potter

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

// A brute force solver for the "Weighted Exact Cover Problem".
package solvers

import (
	"golang.org/x/exp/slices"
	"gonum.org/v1/gonum/stat/combin"
)

// makeSolutionFromSubsets attempts to make an exact cover by adding the candidates (subsets)
// one by one (in order listed in subsetIndicies) until one of the three conditions are met:
// 1. feasible (solution found),
// 2. infeasible (overcovered or undercovered with no subsets left) or
// 3. cost is greater or equal to that in the supplied in `best` if `best.ExactlyCovered`.
//
// It returns a subsetsEval with representing the cover found with ExactlyCovered == true found
// for 1. and ExactlyCovered == false for 2. and 3.
func makeSolutionFromSubsets(ins instance, subsetIndices []int, best subsetsEval) subsetsEval {
	coverCounts := make([]int, ins.n)

	var s subsetsEval

	if ins.n == 0 {
		s.ExactlyCovered = true
		return s
	}

	for _, subsetIdx := range subsetIndices {
		for _, elementIdx := range ins.subsets[subsetIdx] {
			coverCounts[elementIdx] += 1
		}

		s.Cost += ins.costs[subsetIdx]
		// TODO: the use case is find a solution (aka a cover) that is cheaper than
		// a known solution so we could abort the search if s.cost greater than the cost
		// of a known cover.
		s.SubsetsIndices = append(s.SubsetsIndices, subsetIdx)

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
			s.ExactlyCovered = true
			return s
		}

		if best.ExactlyCovered && best.Cost <= s.Cost {
			return s
		}
	}

	return s
}

// SolveByBruteForce attempts finds a minimum cost exact cover for
// an instance.
//
// If a minimum cost exact cover exists, the returned subsetsEval will contain
// indices to this cover and its exactlyCovered flag will be true. Otherwise,
// the zero value of subsetEval will be returned.
func SolveByBruteForce(ins instance) (subsetsEval, error) {

	nSubsetsToTry := ins.n
	// At most len(ins.subsets) are needed because each subset has to cover
	// at least one unique elemnt not covered by the other subsets in an
	// exact cover.
	if len(ins.subsets) < ins.n {
		nSubsetsToTry = len(ins.subsets)
	}

	bestSubsetsEval := makeSolutionFromSubsets(ins, nil, subsetsEval{})

	for i := 1; i <= nSubsetsToTry; i++ {
		combinations := combin.NewCombinationGenerator(len(ins.subsets), i)
		for combinations.Next() {
			perm := combinations.Combination(nil)
			subsetEval := makeSolutionFromSubsets(ins, perm, bestSubsetsEval)
			if !subsetEval.ExactlyCovered {
				continue
			} else if !bestSubsetsEval.ExactlyCovered || subsetEval.Cost < bestSubsetsEval.Cost {
				bestSubsetsEval = subsetEval
			}

		}
	}

	if !bestSubsetsEval.ExactlyCovered {
		return subsetsEval{}, nil
	}

	slices.Sort(bestSubsetsEval.SubsetsIndices)
	return bestSubsetsEval, nil
}
