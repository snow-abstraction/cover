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

package cover

type Instance struct {
	// The number of elements in the set X to be covered, indexed
	// 0 ... N-1.
	N int
	// Subsets of X. The inner slices must only contain element indices in
	// [0, n-1]. The indices must be sorted and each subset must be include
	// at most once. Empty Subsets are not allowed.
	Subsets [][]int
	// The cost of each subset. Each cost must be strictly positive.
	// The length of subsets and Costs must be equal.
	// The restrictions on the Costs reasonable for many problems and
	// suit certain algorithms.
	Costs []float64
}
