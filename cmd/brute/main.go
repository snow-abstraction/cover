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

// A brute force solver for the "Weighted Exact Cover Problem".
package main

import (
	"fmt"

	"github.com/snow-abstraction/cover/internal/solvers"
)

func main() {
	ins, err := solvers.MakeInstance(3, [][]int{{0, 1}, {1, 2}, {0, 2}}, []float64{1.0, 1.0, 1.0})
	if err != nil {
		fmt.Printf("failed to optimal solution due to instance data: %s", err)
	}
	sol, err := solvers.SolveByBruteForce(ins)
	if err != nil {
		fmt.Printf("failed to optimal solution due to error: %s", err)
	}
	fmt.Printf("%+v\n", sol)
}
