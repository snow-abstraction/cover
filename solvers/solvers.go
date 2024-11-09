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

package solvers

import (
	"github.com/snow-abstraction/cover"
	"github.com/snow-abstraction/cover/internal/solvers"
)

// SolveByBranchAndBound attempts finds a minimum cost exact cover for
// an instance by using a branch-and-bound algorithm.
//
// If a minimum cost exact cover exists, the returned subsetsEval will contain
// indices to this cover and its exactlyCovered flag will be true. Otherwise,
// the zero value of subsetEval will be returned.
func SolveByBranchAndBound(ins cover.Instance) (cover.SubsetsEval, error) {
	return solvers.SolveByBranchAndBound(ins, cover.BranchAndBoundConfig{WorkersCount: 1})
}

// SolveByBranchAndBoundWithConfig does what SolveByBranchAndBound does
// except that the extra configuration argument configures the solver.
func SolveByBranchAndBoundWithConfig(
	ins cover.Instance,
	config cover.BranchAndBoundConfig) (cover.SubsetsEval, error) {
	return solvers.SolveByBranchAndBound(ins, config)
}

// SolveByBruteForce attempts finds a minimum cost exact cover for
// an instance by evaluating all possible selections of the subsets.
//
// If a minimum cost exact cover exists, the returned subsetsEval will contain
// indices to this cover and its exactlyCovered flag will be true. Otherwise,
// the zero value of subsetEval will be returned.
func SolveByBruteForce(ins cover.Instance) (cover.SubsetsEval, error) {
	return solvers.SolveByBruteForce(ins)
}
