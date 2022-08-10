// A brute force solver for the "Set Partitioning Problem".
package main

import (
	"fmt"

	"github.com/snow-abstraction/optimal_set_partition/solvers"
)

func main() {
	ins := solvers.MakeInstance(3, [][]int{{0, 1}, {1, 2}, {0, 2}}, []float64{1.0, 1.0, 1.0})
	sol := solvers.SolveByBruteForce(ins)
	fmt.Printf("%+v\n", sol)
}
