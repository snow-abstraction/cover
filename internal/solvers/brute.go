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
	"slices"
)

// updateBestSolutionFromSubsets attempts to make an exact cover by adding the candidates (subsets)
// one by one (in order listed in subsetIndices) until one of the three conditions are met:
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

// SolveByBruteForceInternal attempts finds a minimum cost exact cover for
// an instance by evaluating all possible selections of the subsets.
//
// If a minimum cost exact cover exists, the returned subsetsEval will contain
// indices to this cover and its exactlyCovered flag will be true. Otherwise,
// the zero value of subsetEval will be returned.
func SolveByBruteForceInternal(ins instance) (subsetsEval, error) {
	if ins.m == 0 {
		return subsetsEval{
			ExactlyCovered: true,
			Optimal:        true,
		}, nil
	}

	nSubsetsToTry := ins.m
	// At most len(ins.subsets) are needed because each subset has to cover
	// at least one unique element not covered by the other subsets in an
	// exact cover.
	if len(ins.subsets) < ins.m {
		nSubsetsToTry = len(ins.subsets)
	}

	subsetsScratch := make([]int, 0, nSubsetsToTry)
	coverCountsScratch := make([]int, ins.m)

	var bestSubsetsEval subsetsEval
	bestSubsetsEval.SubsetsIndices = make([]int, 0, nSubsetsToTry)
	for i := 1; i <= nSubsetsToTry; i++ {
		combinations := newCombinationGenerator(len(ins.subsets), i)
		for combinations.Next() {
			updateBestSolutionFromSubsets(ins, combinations.combination, subsetsScratch, coverCountsScratch, &bestSubsetsEval)
		}
	}

	if !bestSubsetsEval.ExactlyCovered {
		return subsetsEval{}, nil
	}

	slices.Sort(bestSubsetsEval.SubsetsIndices)
	bestSubsetsEval.Optimal = true
	return bestSubsetsEval, nil
}

type CombinationGenerator struct {
	n           int
	k           int
	combination []int
}

// Make a generator generating 0-indexed (n, k) combinations in lexicographical order starting with
// 0, 1, ..., k-1 and ending with n - k, ..., n -1
func newCombinationGenerator(n, k int) *CombinationGenerator {
	return &CombinationGenerator{n, k, nil}
}

func (c *CombinationGenerator) Next() bool {
	if c.combination == nil {
		if c.n < 1 || c.k < 1 {
			return false
		}
		if c.n < c.k {
			return false
		}
		initialCombination := make([]int, c.k)
		for i := 0; i < c.k; i++ {
			initialCombination[i] = i
		}
		c.combination = initialCombination
		return true
	}

	for i := c.k - 1; i >= 0; i-- {
		if c.combination[i] < c.n-c.k+i {
			c.combination[i]++
			for j := i + 1; j < c.k; j++ {
				c.combination[j] = c.combination[j-1] + 1
			}
			return true
		}
	}
	return false

}
