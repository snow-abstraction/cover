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

func convertSubsetsToMatrix(subsets [][]int) (cCSMatrix, error) {
	x := make([]uint32, 0)

	for i := 0; i < len(subsets); i++ {
		subset := subsets[i]
		for j := 0; j < len(subset); j++ {
			x = append(x, uint32(subset[j]))
		}
		x = append(x, sen)
	}

	return makeCompressedColumnMatrix(x)
}
