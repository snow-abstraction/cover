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

import "github.com/snow-abstraction/cover"

// SolveByBranchAndBound exposes an internal method without the suffix `Internal“
// and takes and returns exported types.
func SolveByBranchAndBound(ins cover.Instance) (cover.SubsetsEval, error) {
	solverInstance, err := MakeInstance(ins.ElementCount, ins.Subsets, ins.Costs)
	if err != nil {
		return cover.SubsetsEval{}, err
	}

	sol, err := SolveByBranchAndBoundInternal(solverInstance)
	return cover.SubsetsEval(sol), err
}

// SolveByBruteForce exposes an internal method without the suffix `Internal“
// and takes and returns exported types.
func SolveByBruteForce(ins cover.Instance) (cover.SubsetsEval, error) {
	solverInstance, err := MakeInstance(ins.ElementCount, ins.Subsets, ins.Costs)
	if err != nil {
		return cover.SubsetsEval{}, err
	}

	sol, err := SolveByBruteForceInternal(solverInstance)
	return cover.SubsetsEval(sol), err
}
