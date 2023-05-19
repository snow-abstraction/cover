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

package main

import (
	"math"

	"github.com/snow-abstraction/cover/tree"
)

func main() {

	unprocessed := tree.CreateInitialNodes()
	k3, k4 := unprocessed[0].Branch(math.MaxFloat64, 3, 4)
	unprocessed = append(unprocessed, k3, k4)

	// for len(unprocessed) > 0 {

	// 	node := unprocessed[0]
	// 	unprocessed[0] = nil
	// 	unprocessed = unprocessed[1:]

	// 	// find lb
	// 	// check if integral solution found
	// 	// find branching rows/constraints
	// 	// prune or branch

	// 		printTree([]*Node{node})

	// }

	tree.PrintTree(unprocessed)
}
