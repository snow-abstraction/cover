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

// A brute force solver for the "Weighted Exact Cover Problem".
package solvers

import (
	"golang.org/x/exp/slices"
	"gonum.org/v1/gonum/stat/combin"
)

// updateBestSolutionFromSubsets attempts to make an exact cover by adding the candidates (subsets)
// one by one (in order listed in subsetIndicies) until one of the three conditions are met:
// 1. feasible (solution found),
// 2. infeasible (overcovered or undercovered with no subsets left) or
// 3. cost is greater or equal to that in the supplied in `best` if `best.ExactlyCovered`.
//
// If the solution found is cheaper than `best.Cost` then `best` is updated.
//
// Note: *Scratch arguments are used for performance by avoiding garbage collector work.
func updateBestSolutionFromSubsets(ins instance, subsetIndices []int, subsetsScratch []int, coverCountsScratch []int, best *subsetsEval) {

	// reset scratch
	subsetsScratch = subsetsScratch[:0]
	for i := 0; i < len(coverCountsScratch); i++ {
		coverCountsScratch[i] = 0
	}

	cost := 0.0
	for _, subsetIdx := range subsetIndices {
		for _, elementIdx := range ins.subsets[subsetIdx] {
			(coverCountsScratch)[elementIdx] += 1
		}

		cost += ins.costs[subsetIdx]
		if best.ExactlyCovered && cost >= best.Cost {
			return
		}

		subsetsScratch = append(subsetsScratch, subsetIdx)

		allConstraintsCoveredExactly := true
		for _, coverCount := range coverCountsScratch {
			isOverCovered := coverCount > 1
			if isOverCovered {
				return
			} else if coverCount == 0 {
				allConstraintsCoveredExactly = false
			}
		}

		if allConstraintsCoveredExactly {
			if !best.ExactlyCovered || cost < best.Cost {
				best.Cost = cost
				best.ExactlyCovered = true
				best.SubsetsIndices = best.SubsetsIndices[:len(subsetsScratch)]
				copy(best.SubsetsIndices, subsetsScratch)
				// cannot improve an exact cover by adding subsets
				return
			}
		}
	}
}

// SolveByBruteForce attempts finds a minimum cost exact cover for
// an instance.
//
// If a minimum cost exact cover exists, the returned subsetsEval will contain
// indices to this cover and its exactlyCovered flag will be true. Otherwise,
// the zero value of subsetEval will be returned.
func SolveByBruteForce(ins instance) (subsetsEval, error) {
	if ins.m == 0 {
		return subsetsEval{
			ExactlyCovered: true,
		}, nil
	}

	nSubsetsToTry := ins.m
	// At most len(ins.subsets) are needed because each subset has to cover
	// at least one unique elemnt not covered by the other subsets in an
	// exact cover.
	if len(ins.subsets) < ins.m {
		nSubsetsToTry = len(ins.subsets)
	}

	subsetsScratch := make([]int, 0, nSubsetsToTry)
	coverCountsScratch := make([]int, ins.m)

	var bestSubsetsEval subsetsEval
	bestSubsetsEval.SubsetsIndices = make([]int, 0, nSubsetsToTry)
	comb := make([]int, 0, len(ins.subsets))
	for i := 1; i <= nSubsetsToTry; i++ {
		combinations := combin.NewCombinationGenerator(len(ins.subsets), i)
		comb = comb[:i]
		for combinations.Next() {
			combinations.Combination(comb)
			updateBestSolutionFromSubsets(ins, comb, subsetsScratch, coverCountsScratch, &bestSubsetsEval)
		}
	}

	if !bestSubsetsEval.ExactlyCovered {
		return subsetsEval{}, nil
	}

	slices.Sort(bestSubsetsEval.SubsetsIndices)
	return bestSubsetsEval, nil
}
