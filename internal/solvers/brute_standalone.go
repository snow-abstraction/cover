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
	"fmt"
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
func updateBestSolutionFromSubsets2(ins instance, combin uint32, subsetsScratch []int, coverCountsScratch []int, best *subsetsEval) {

	// reset scratch
	subsetsScratch = subsetsScratch[:0]
	for i := 0; i < len(coverCountsScratch); i++ {
		coverCountsScratch[i] = 0
	}

	cost := 0.0
	for subsetIdx := range len(ins.subsets) {
		if (combin & (uint32(1) << subsetIdx)) == 0 {
			continue
		}

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
//
// Note: we resist implementing performance improving optimizations to stick
// a simple brute force implementation. The slow implementation means that
// uint32 is sufficient for representing solutions.
func SolveByBruteForceInternal2(ins instance) (subsetsEval, error) {
	n := len(ins.subsets)
	if n > 32 {
		return subsetsEval{}, fmt.Errorf("the instance has %d subsets but the brute solver supports at most 32", n)
	}

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
	if n < ins.m {
		nSubsetsToTry = n
	}

	subsetsScratch := make([]int, 0, nSubsetsToTry)
	coverCountsScratch := make([]int, ins.m)

	var bestSubsetsEval subsetsEval
	bestSubsetsEval.SubsetsIndices = make([]int, 0, nSubsetsToTry)

	// comb is short for combination and is bitset representing
	// which subsets are chosen. Initially subset 0 is chosen.
	comb := uint32(1)
	// last is the last bit to try and represents choosing all subsets.
	var last uint32
	for i := 0; i < n; i += 1 {
		last |= (uint32(1) << i)
	}

	// the form of the loop is prevent overflow
	for {
		updateBestSolutionFromSubsets2(ins, comb, subsetsScratch, coverCountsScratch, &bestSubsetsEval)
		if comb == last {
			break
		}
		comb += 1
	}

	if !bestSubsetsEval.ExactlyCovered {
		return subsetsEval{}, nil
	}

	slices.Sort(bestSubsetsEval.SubsetsIndices)
	bestSubsetsEval.Optimal = true
	return bestSubsetsEval, nil
}
